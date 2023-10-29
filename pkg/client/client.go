package client

import (
	"encoding/json"
	"fmt"
	"log"
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

	User  *auth.User `json:"user"`           // user
	Root  string     `json:"root"`           // path to root drive for users files and directories
	SfDir string     `json:"state_file_dir"` // path to state file

	Monitor *monitor.Monitor `json:"monitor"` // listener that checks for file or directory events
	Drive   *svc.Drive       `json:"drive"`   // client drive for managing users files and directories

	Db *db.Query `json:"db"` // local db connection

	Handlers map[string]EHandler // map of active event handlers for individual files
	Transfer *transfer.Transfer  // file transfer component
}

func NewClient(user, userID string) *Client {
	conf := ClientConfig()

	// set up local client services
	svcRoot := filepath.Join(conf.Root, user)
	root := svc.NewDirectory("root", conf.User, svcRoot)
	drv := svc.NewDrive(auth.NewUUID(), conf.User, conf.User, root.Path, root)

	// // set up monitor
	// monitor := monitor.NewMonitor(drv.Root.Path)
	// if err := monitor.Start(drv.Root.Path); err != nil {
	// 	log.Fatalf("failed to set up file listeners: %v", err)
	// }

	return &Client{
		StartTime: time.Now().UTC(),
		Conf:      conf,
		User:      auth.NewUser(user, userID, conf.Email, auth.NewUUID(), svcRoot, false),
		Root:      filepath.Join(svcRoot, "root"),
		SfDir:     filepath.Join(svcRoot, "state"),
		// Monitor:   monitor,
		Drive: drv,
		Db:    db.NewQuery(filepath.Join(svcRoot, "dbs"), true),
		// Handlers:  make(map[string]EHandler),
		// Transfer:  transfer.NewTransfer(),
	}
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

func (c *Client) RemoveUser(userID string) error {
	if c.User == nil {
		return fmt.Errorf("no user (id=%s) found", userID)
	} else if c.User.ID == userID {
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
