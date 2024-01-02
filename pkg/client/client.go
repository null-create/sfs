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

	Tok *auth.Token `json:"token"` // token creator for requests

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

	// map of handler off switches. used during shutdown.
	OffSwitches map[string]chan bool `json:"-"`

	// file transfer component. handles file uploads and downloads.
	Transfer *transfer.Transfer `json:"-"`

	// http client. used mainly for small-scope calls to the server.
	Client *http.Client `json:"-"`
}

// build services endpoints map.
// individual files and directories have endpoints defined
// within their respective data structures.
func (c *Client) setEndpoints() {
	EndpointRootWithPort := fmt.Sprint(EndpointRoot, ":", c.Conf.Port)

	c.Endpoints["files"] = fmt.Sprint(EndpointRootWithPort, "v1/files/all")
	c.Endpoints["new file"] = fmt.Sprint(EndpointRootWithPort, "v1/files/new")
	c.Endpoints["dirs"] = fmt.Sprint(EndpointRootWithPort, "/v1/dirs")
	c.Endpoints["new dir"] = fmt.Sprintf(EndpointRootWithPort, "/dirs/new")
	c.Endpoints["drive"] = fmt.Sprint(EndpointRootWithPort, "/v1/drive/", c.Drive.ID)
	c.Endpoints["new drive"] = fmt.Sprint(EndpointRootWithPort, "/v1/drive/new")
	c.Endpoints["sync"] = fmt.Sprint(EndpointRootWithPort, "/v1/sync/", c.Drive.ID)
	c.Endpoints["user"] = fmt.Sprint(EndpointRootWithPort, "/v1/users/", c.User.ID)
	c.Endpoints["new user"] = fmt.Sprint(EndpointRootWithPort, "/v1/users/new")
	c.Endpoints["all users"] = fmt.Sprint(EndpointRootWithPort, "/v1/users/all")
}

// creates a new client object. does not create actual service directories or
// other necessary infrastructure -- only the client itself.
func NewClient(user *auth.User) *Client {
	// get client service configs
	cfg := ClientConfig()

	// set up local client services
	driveID := auth.NewUUID()
	svcRoot := filepath.Join(cfg.Root, cfg.User)
	root := svc.NewRootDirectory("root", cfg.User, driveID, filepath.Join(svcRoot, "root"))
	drv := svc.NewDrive(driveID, cfg.User, user.ID, root.Path, root.ID, root)

	// intialize client
	c := &Client{
		StartTime:   time.Now().UTC(),
		Conf:        cfg,
		UserID:      user.ID,
		User:        user,
		Root:        filepath.Join(svcRoot, "root"),
		SfDir:       filepath.Join(svcRoot, "state"),
		Endpoints:   make(map[string]string),
		Monitor:     monitor.NewMonitor(drv.Root.Path),
		Off:         make(chan bool),
		Drive:       drv,
		Db:          db.NewQuery(filepath.Join(svcRoot, "dbs"), true),
		Handlers:    make(map[string]EHandler),
		OffSwitches: make(map[string]chan bool),
		Transfer:    transfer.NewTransfer(cfg.Port),
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// run discover to populate the database and internal data structures
	// with users files and directories
	root, err := c.Discover(root)
	if err != nil {
		log.Fatalf("failed to discover user file system: %v", err)
	}
	drv.Root = root
	drv.IsLoaded = true

	// add drive itself to DB (root was added during discovery)
	// then attach to client
	if err := c.Db.AddDrive(c.Drive); err != nil {
		log.Fatal(fmt.Errorf("failed to add client drive to database: %v", err))
	}
	c.Drive = drv
	if err := c.Drive.SaveState(); err != nil {
		log.Fatalf("failed to save drive state: %v", err)
	}

	// build services endpoints map (files and directories have endpoints defined
	// within their respective data structures)
	c.setEndpoints()

	// add token componet
	c.Tok = auth.NewT()

	// start monitoring services
	if err := c.Monitor.Start(root.Path); err != nil {
		log.Fatal("failed to start monitor", err)
	}

	// build and start monitoring event handlers
	c.BuildHandlers()
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
	return os.WriteFile(filepath.Join(c.SfDir, fn), data, svc.PERMS)
}

// shutdown client side services
func (c *Client) ShutDown() error {
	// stop sync doc monitoring loop
	c.Off <- true

	// shut down monitor and handlers
	if err := c.StopMonitoring(); err != nil {
		return fmt.Errorf("failed to shutdown monitor: %v", err)
	}
	c.StopHandlers()

	if err := c.SaveState(); err != nil {
		return fmt.Errorf("failed to save state: %v", err)
	}
	return nil
}

// start up client services
func (c *Client) Start() error {
	if !c.Drive.IsLoaded {
		if c.Drive.Root.IsNil() || c.Drive.Root.IsEmpty() {
			root, err := c.Db.GetDirectory(c.Drive.RootID)
			if err != nil {
				return err
			}
			c.Drive.Root = c.Populate(root)
		}
		c.Drive.IsLoaded = true
	}

	// start monitoring services
	if err := c.Monitor.Start(c.Drive.Root.Path); err != nil {
		log.Fatal("failed to start monitor", err)
	}

	// start monitoring event handlers
	if err := c.StartHandlers(); err != nil {
		log.Fatal("failed to start event handlers", err)
	}

	// start sync doc monitoring
	c.Sync(c.Off)

	// save initial state
	if err := c.SaveState(); err != nil {
		log.Fatal("failed to save state")
	}

	return nil
}
