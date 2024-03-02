package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/db"
	"github.com/sfs/pkg/monitor"
	svc "github.com/sfs/pkg/service"
	"github.com/sfs/pkg/transfer"
)

// Client-side SFS service instance.
type Client struct {
	StartTime  time.Time  `json:"start_time"`      // start time for this client
	Conf       *Conf      `json:"client_settings"` // client service settings
	User       *auth.User `json:"user"`            // user object
	UserID     string     `json:"user_id"`         // usersID for this client
	DriveID    string     `json:"drive_id"`        // drive ID for this client
	Root       string     `json:"root"`            // path to root sfs directory for users files and directories
	SfDir      string     `json:"state_file_dir"`  // path to state file
	RecycleBin string     `json:"recycle_bin"`     // path to recycle bin. "deleted" items live here.
	Drive      *svc.Drive `json:"drive"`           // client drive for managing users files and directories
	Db         *db.Query  `json:"db"`              // local db connection

	// token creator for requests
	Tok *auth.Token `json:"token"`

	// server api endpoints.
	// file objects have their own API field, this is for storing
	// general operation endpoints like sync operations.
	// key == file or dir uuid, val == associated server API endpoint
	Endpoints map[string]string `json:"endpoints"`

	// listener that checks for file or directory events
	Monitor *monitor.Monitor `json:"-"`

	// map of active event handlers for individual files and directories
	// key == item path, value == event handler function
	Handlers map[string]func() `json:"-"`

	// map of listener off switches. used during shutdown.
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
func (c *Client) ShutDown() {
	c.StopMonitoring()
	c.StopHandlers()

	if err := c.SaveState(); err != nil {
		log.Printf("[ERROR] failed to save state file: %v", err)
	}
}

// start up client services. creates a blocking process
// to facilitate monitoring and synchronization services
//
// NOTE: the shutDown parameter is used for programmatic
// testing purposes. clients are normally shut down using ctrl-c.
func (c *Client) start(shutDown chan os.Signal) error {
	// NOTE: this check may be redundant since start is always called
	// after LoadClient().
	if !c.Drive.IsLoaded || c.Drive.Root.IsEmpty() {
		if err := c.LoadDrive(); err != nil {
			return err
		}
		// refresh drive on startup. this will find anything that was
		// added if SFS wasn't running at the time.
		c.RefreshDrive()
	}
	// save initial state
	if err := c.SaveState(); err != nil {
		return fmt.Errorf("failed to save initial state: %v", err)
	}
	// wait for signal (such as ctrl-c or some other syscall) to shutdown client.
	// we want to make start a blocking process so all the goroutines
	// that are monitoring files (and all their event listeners)
	// can actually run.
	<-shutDown

	// "gracefully" shutdown when we receive a signal.
	c.ShutDown()
	return nil
}

// start sfs client service. creates a blocking process
// to allow monitoring and synchronization services to run.
// assumes that LoadClient(true) has already been called prior to being called.
// sadness will follow if this is not the case.
func (c *Client) Start() error {
	shutDown := make(chan os.Signal)
	signal.Notify(shutDown, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	if err := c.start(shutDown); err != nil {
		log.Fatal(err)
	}
	return nil
}
