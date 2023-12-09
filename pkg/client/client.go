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
	// get configs
	conf := ClientConfig()

	// set up local client services
	svcRoot := filepath.Join(conf.Root, conf.User)
	root := svc.NewRootDirectory("root", conf.User, filepath.Join(svcRoot, "root"))
	drv := svc.NewDrive(auth.NewUUID(), conf.User, user.ID, root.Path, root.ID, root)

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

	// build services endpoints map
	EndpointRootWithPort := fmt.Sprint(EndpointRoot, ":", c.Conf.Port)
	files := c.Drive.GetFiles()
	for _, f := range files {
		c.Endpoints[f.ID] = f.Endpoint
	}
	subDirs := c.Drive.GetDirs()
	for _, d := range subDirs {
		c.Endpoints[d.ID] = d.Endpoint
	}
	c.Endpoints["drive"] = fmt.Sprint(EndpointRootWithPort, "/v1/drive/", c.Drive.ID)
	c.Endpoints["sync"] = fmt.Sprint(EndpointRootWithPort, "/v1/sync/", c.Drive.ID)
	c.Endpoints["root"] = EndpointRootWithPort

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
