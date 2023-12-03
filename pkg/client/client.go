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
	// key == fileID, value is the associated endpoint
	Endpoints map[string]string `json:"endpoints"`

	// listener that checks for file or directory events
	Monitor *monitor.Monitor `json:"-"` // json ignore tags

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
	// get configs
	conf := ClientConfig()

	// set up local client services
	svcRoot := filepath.Join(conf.Root, conf.User)
	root := svc.NewDirectory("root", conf.User, filepath.Join(svcRoot, "root"))
	drv := svc.NewDrive(auth.NewUUID(), conf.User, user.ID, root.Path, root.ID, root)

	// intialize client and start monitoring service
	c := &Client{
		StartTime: time.Now().UTC(),
		Conf:      conf,
		UserID:    user.ID,
		User:      user,
		Root:      filepath.Join(svcRoot, "root"),
		SfDir:     filepath.Join(svcRoot, "state"),
		Monitor:   monitor.NewMonitor(drv.Root.Path),
		Drive:     drv,
		Db:        db.NewQuery(filepath.Join(svcRoot, "dbs"), true),
		Handlers:  make(map[string]EHandler),
		Transfer:  transfer.NewTransfer(),
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	c.BuildHandlers()
	if err := c.Monitor.Start(root.Path); err != nil {
		log.Fatal("failed to start monitor", err)
	}
	if err := c.StartHandlers(); err != nil {
		log.Fatal("failed to start event handlers", err)
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

// ---- user functions --------------------------------

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
			return nil, fmt.Errorf("failed to get user from database")
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
		c.Drive.Remove()
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

// ----- files --------------------------------------

// add a file to a specified directory
func (c *Client) AddFile(dirID string, file *svc.File) error {
	return c.Drive.AddFile(dirID, file)
}

// add a series of files to a specified directory
func (c *Client) AddFiles(dirID string, files []*svc.File) error {
	if len(files) == 0 {
		return fmt.Errorf("no files to add")
	}
	for _, file := range files {
		if err := c.Drive.AddFile(dirID, file); err != nil {
			return err
		}
	}
	return nil
}

// update a file in a specied directory
func (c *Client) UpdateFile(dirID string, fileID string, data []byte) error {
	file := c.Drive.GetFile(fileID)
	if file == nil {
		return fmt.Errorf("no file ID %s found", fileID)
	}
	if len(data) == 0 {
		return fmt.Errorf("no data received")
	}
	return c.Drive.UpdateFile(dirID, file, data)
}

// remove a file in a specied directory
func (c *Client) RemoveFile(dirID string, file *svc.File) error {
	return c.Drive.RemoveFile(dirID, file)
}

// remove files from a specied directory
func (c *Client) RemoveFiles(dirID string, fileIDs []string) error {
	if len(fileIDs) == 0 {
		return fmt.Errorf("no files to remove")
	}
	for _, fileID := range fileIDs {
		file := c.Drive.GetFile(fileID)
		if file == nil { // didn't find the file. ignore.
			continue
		}
		if err := c.Drive.RemoveFile(dirID, file); err != nil {
			return err
		}
	}
	return nil
}

// ----- directories --------------------------------

func (c *Client) AddDir(dir *svc.Directory) error {
	return c.Drive.AddDir(dir)
}

func (c *Client) AddDirs(dirs []*svc.Directory) error {
	return c.Drive.AddDirs(dirs)
}

func (c *Client) RemoveDir(dirID string) error {
	return c.Drive.RemoveDir(dirID)
}

func (c *Client) RemoveDirs(dirs []*svc.Directory) error {
	return c.Drive.RemoveDirs(dirs)
}

// ----- drive --------------------------------

func (c *Client) GetDrive(driveID string) (*svc.Drive, error) {
	if c.Drive == nil {
		log.Print("[WARNING] drive not found! attempting to load...")
		drive, err := c.Db.GetDrive(driveID)
		if err != nil {
			return nil, err
		}
		root, err := c.Db.GetDirectory(drive.RootID)
		if err != nil {
			return nil, err
		}
		drive.Root = root
		if drive.SyncIndex == nil {
			drive.SyncIndex = drive.Root.WalkS(svc.NewSyncIndex(drive.OwnerID))
		}
		c.Drive = drive
	}
	return c.Drive, nil
}
