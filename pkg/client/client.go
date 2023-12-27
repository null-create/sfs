package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/db"
	"github.com/sfs/pkg/monitor"
	svc "github.com/sfs/pkg/service"
	"github.com/sfs/pkg/transfer"
)

type Client struct {
	StartTime time.Time `json:"start_time"`      // start time for this client
	Conf      *Conf     `json:"client_settings"` // client service settings

	User   *auth.User `json:"user"`           // user object
	UserID string     `json:"user_id"`        // usersID for this client
	Root   string     `json:"root"`           // path to root drive for users files and directories
	SfDir  string     `json:"state_file_dir"` // path to state file

	Drive *svc.Drive `json:"drive"` // client drive for managing users files and directories
	Db    *db.Query  `json:"db"`    // local db connection

	// server api endpoints.
	// file objects have their own API field, this is for storing
	// general operation endpoints like sync operations
	// key == file or dir uuid, val == associated server API endpoint
	Endpoints map[string]string `json:"endpoints"`

	// listener that checks for file or directory events
	Monitor *monitor.Monitor `json:"-"`

	// Channel for managing sync loop
	Off chan bool `json:"-"`

	// map of active event handlers for individual files
	// key == filepath, value == new EventHandler() function
	Handlers map[string]EHandler `json:"-"`

	// file transfer component
	Transfer *transfer.Transfer `json:"-"`

	// http client
	Client *http.Client `json:"-"`
}

// creates a new client object. does not create actual service directories or
// other necessary infrastructure -- only the client itself.
func NewClient(user *auth.User) *Client {
	// get client service configs
	conf := ClientConfig()

	// set up local client services
	driveID := auth.NewUUID()
	svcRoot := filepath.Join(conf.Root, conf.User)
	root := svc.NewRootDirectory("root", conf.User, driveID, filepath.Join(svcRoot, "root"))
	drv := svc.NewDrive(driveID, conf.User, user.ID, root.Path, root.ID, root)

	// intialize client
	c := &Client{
		StartTime: time.Now().UTC(),
		Conf:      conf,
		UserID:    user.ID,
		User:      user,
		Root:      filepath.Join(svcRoot, "root"),
		SfDir:     filepath.Join(svcRoot, "state"),
		Endpoints: make(map[string]string),
		Monitor:   monitor.NewMonitor(drv.Root.Path),
		Off:       make(chan bool),
		Drive:     drv,
		Db:        db.NewQuery(filepath.Join(svcRoot, "dbs"), true),
		Handlers:  make(map[string]EHandler),
		Transfer:  transfer.NewTransfer(conf.Port),
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// run discover to populate the database and internal data structures
	// with users files and directories
	drv.Root = c.Discover(root)
	drv.IsLoaded = true

	// add drive itself to DB
	if err := c.Db.AddDrive(c.Drive); err != nil {
		log.Fatal(fmt.Errorf("failed to add client drive to database: %v", err))
	}

	// build services endpoints map (files and directories have endpoints defined
	// within their respective data structures)
	EndpointRootWithPort := fmt.Sprint(EndpointRoot, ":", c.Conf.Port)
	c.Endpoints["drive"] = fmt.Sprint(EndpointRootWithPort, "/v1/drive/", c.Drive.ID)
	c.Endpoints["sync"] = fmt.Sprint(EndpointRootWithPort, "/v1/sync/", c.Drive.ID)

	// build and start handlers
	c.BuildHandlers()
	if err := c.Monitor.Start(root.Path); err != nil {
		log.Fatal("failed to start monitor", err)
	}
	if err := c.StartHandlers(); err != nil {
		log.Fatal("failed to start event handlers", err)
	}

	// start sync doc monitoring
	c.Sync(c.Off)

	// save initial state
	if err := c.SaveState(); err != nil {
		log.Fatal("failed to save state")
	}
	return c
}

// remove previous state file(s)
func (c *Client) cleanSfDir() error {
	entries, err := os.ReadDir(c.SfDir)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}
	for _, entry := range entries {
		if err := os.Remove(filepath.Join(c.SfDir, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

// save client state
func (c *Client) SaveState() error {
	if err := c.cleanSfDir(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode json: %v", err)
	}
	fn := fmt.Sprintf("client-state-%s.json", time.Now().UTC().Format("2006-01-02T15-04-05Z"))
	fp := filepath.Join(c.SfDir, fn)
	return os.WriteFile(fp, data, svc.PERMS)
}

// ------- user functions --------------------------------

func (c *Client) AddUser(user *auth.User) error {
	if c.User == nil {
		c.User = user
	} else {
		return fmt.Errorf("cannot have more than one user: %v", c.User)
	}
	if err := c.SaveState(); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetUser() (*auth.User, error) {
	if c.User == nil {
		log.Print("[WARNING] client instance has no user object. attempting to get user from the database...")
		if c.Db == nil {
			return nil, fmt.Errorf("failed to get user. database not initialized")
		}
		user, err := c.Db.GetUser(c.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user from database: %v", err)
		}
		return user, nil
	}
	return c.User, nil
}

func (c *Client) UpdateUser(user *auth.User) error {
	if user.ID == c.User.ID {
		c.User = user
	} else {
		return fmt.Errorf("user (id=%s) is not client user (id=%s)", user.ID, c.User.ID)
	}
	if err := c.SaveState(); err != nil {
		return err
	}
	return nil
}

// remove a user and their drive from the client instance.
// clears all users files and directores, as well as removes the
// user from the client instance and db.
func (c *Client) RemoveUser(userID string) error {
	if c.User == nil {
		return fmt.Errorf("no user (id=%s) found", userID)
	} else if c.User.ID == userID {
		// remove drive and users files
		c.Drive.ClearDrive()
		// remove user and info from database
		if err := c.Db.RemoveUser(c.UserID); err != nil {
			return err
		}
		c.User = nil
		log.Printf("[INFO] user %s removed", userID)
	} else {
		return fmt.Errorf("wrong user ID (id=%s)", userID)
	}
	if err := c.SaveState(); err != nil {
		return err
	}
	return nil
}
