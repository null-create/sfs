package client

import (
	"encoding/json"
	"fmt"
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

	DriveID string     `json:"drive_id"` // drive ID for this client
	Drive   *svc.Drive `json:"drive"`    // client drive for managing users files and directories
	Db      *db.Query  `json:"db"`       // local db connection

	Tok *auth.Token `json:"token"` // token creator for requests

	// server api endpoints.
	// file objects have their own API field, this is for storing
	// general operation endpoints like sync operations.
	// key == file or dir uuid, val == associated server API endpoint
	Endpoints map[string]string `json:"endpoints"`

	// listener that checks for file or directory events
	Monitor *monitor.Monitor `json:"-"`

	// map of active event handlers for individual files
	// key == filepath, value == new event hander function
	Handlers map[string]func() `json:"-"`

	// map of handler off switches. used during shutdown.
	// key == filepath, value == chan bool
	OffSwitches map[string]chan bool `json:"-"`

	// file transfer component. handles file uploads and downloads.
	Transfer *transfer.Transfer `json:"-"`

	// http client. used mainly for small-scope calls to the server.
	Client *http.Client `json:"-"`
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
	// shut down monitor and handlers
	c.StopMonitoring()
	c.StopHandlers()

	if err := c.SaveState(); err != nil {
		return fmt.Errorf("failed to save state: %v", err)
	}
	return nil
}

// start up client services.
// TODO: this needs to have some kind of CLI event loop that
// allows the program to persist and recieve commands from the user,
// like how the server persists when listening on the network.
func (c *Client) Start() error {
	if !c.Drive.IsLoaded || c.Drive.Root.IsEmpty() {
		root, err := c.Db.GetDirectory(c.Drive.RootID)
		if err != nil {
			return fmt.Errorf("failed to get root directory: %v", err)
		}
		if root == nil {
			return fmt.Errorf("no root directory found for drive (id=%s)", c.Drive.ID)
		}
		c.Drive.Root = c.Populate(root)
		c.Drive.IsLoaded = true
	}

	// start monitoring services
	if err := c.Monitor.Start(c.Drive.Root.Path); err != nil {
		return fmt.Errorf("failed to start monitoring: %v", err)
	}

	// start monitoring event handlers
	if err := c.StartHandlers(); err != nil {
		return fmt.Errorf("failed to start event handlers: %v", err)
	}

	// save initial state
	if err := c.SaveState(); err != nil {
		return fmt.Errorf("failed to save initial state: %v", err)
	}
	return nil
}
