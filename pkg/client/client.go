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

// build services endpoints map.
// individual files and directories have endpoints defined
// within their respective data structures.
func (c *Client) setEndpoints() {
	EndpointRootWithPort := fmt.Sprint(EndpointRoot, ":", c.Conf.Port)
	// general purpose endpoints.
	// files and directories have their endpoints defined in their respective structures.
	c.Endpoints["all files"] = EndpointRootWithPort + "/v1/files/i/all/" + c.UserID
	c.Endpoints["new file"] = EndpointRootWithPort + "/v1/files/new"
	c.Endpoints["all dirs"] = EndpointRootWithPort + "/v1/i/dirs/all/" + c.UserID
	c.Endpoints["new dir"] = EndpointRootWithPort + "/v1/dirs/new"
	c.Endpoints["drive"] = EndpointRootWithPort + "/v1/drive/" + c.DriveID
	c.Endpoints["new drive"] = EndpointRootWithPort + "/v1/drive/new"
	c.Endpoints["sync"] = EndpointRootWithPort + "/v1/sync/" + c.DriveID
	c.Endpoints["get index"] = EndpointRootWithPort + "/v1/sync/" + c.DriveID
	c.Endpoints["gen index"] = EndpointRootWithPort + "/v1/sync/index/" + c.DriveID + "/index"
	c.Endpoints["gen updates"] = EndpointRootWithPort + "/v1/sync/update/" + c.DriveID + "/update"
	c.Endpoints["user"] = EndpointRootWithPort + "/v1/users/" + c.UserID
	c.Endpoints["new user"] = EndpointRootWithPort + "/v1/users/new"
	c.Endpoints["all users"] = EndpointRootWithPort + "/v1/users/all"
}

// creates a new client object. does not create actual service directories or
// other necessary infrastructure -- only the client itself.
func NewClient(user *auth.User) (*Client, error) {
	ccfg := ClientConfig()

	// set up local client services
	driveID := auth.NewUUID()
	svcRoot := filepath.Join(ccfg.Root, ccfg.User)
	root := svc.NewRootDirectory("root", ccfg.User, driveID, filepath.Join(svcRoot, "root"))
	drv := svc.NewDrive(driveID, ccfg.User, user.ID, root.Path, root.ID, root)
	user.DriveID = driveID
	user.DrvRoot = drv.DriveRoot
	user.SvcRoot = root.Path

	// intialize client
	c := &Client{
		StartTime:   time.Now().UTC(),
		Conf:        ccfg,
		UserID:      user.ID,
		User:        user,
		Root:        filepath.Join(svcRoot, "root"),
		SfDir:       filepath.Join(svcRoot, "state"),
		Endpoints:   make(map[string]string),
		Monitor:     monitor.NewMonitor(drv.Root.Path),
		DriveID:     driveID,
		Drive:       drv,
		Db:          db.NewQuery(filepath.Join(svcRoot, "dbs"), true),
		Handlers:    make(map[string]func()),
		OffSwitches: make(map[string]chan bool),
		Transfer:    transfer.NewTransfer(ccfg.Port),
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// run discover to populate the database and internal data structures
	// with users files and directories
	root, err := c.Discover(root)
	if err != nil {
		return nil, fmt.Errorf("failed to discover user file system: %v", err)
	}
	drv.Root = root
	drv.IsLoaded = true

	// add drive itself to DB (root was added during discovery)
	// then attach to client
	if err := c.Db.AddDrive(drv); err != nil {
		return nil, fmt.Errorf("failed to add client drive to database: %v", err)
	}
	if err := c.Db.AddDir(root); err != nil {
		return nil, fmt.Errorf("failed to add root directory to database: %v", err)
	}
	c.Drive = drv

	// build services endpoints map (files and directories have endpoints defined
	// within their respective data structures)
	c.setEndpoints()

	// add token component
	c.Tok = auth.NewT()

	// start monitoring services
	if err := c.Monitor.Start(root.Path); err != nil {
		return nil, fmt.Errorf("failed to start monitor: %v", err)
	}

	// build and start monitoring event handlers
	if err := c.BuildHandlers(); err != nil {
		return nil, fmt.Errorf("failed to build event handlers: %v", err)
	}
	if err := c.StartHandlers(); err != nil {
		return nil, fmt.Errorf("failed to start event handlers: %v", err)
	}

	// save initial state
	if err := c.SaveState(); err != nil {
		return nil, fmt.Errorf("failed to save initial state: %v", err)
	}
	return c, nil
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

// start up client services
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
