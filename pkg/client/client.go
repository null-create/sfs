package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sfs/pkg/db"
	svc "github.com/sfs/pkg/service"
)

type Client struct {
	StartTime time.Time `json:"start_time"`      // start time for this client
	Conf      *Conf     `json:"client_settings"` // client service settings

	User   string `json:"user"`           // user name
	UserID string `json:"user_id"`        // user ID
	SfDir  string `json:"state_file_dir"` // path to state file

	Db     *db.Query `json:"db"` // local db connection
	client *http.Client
}

func NewClient(user, userID string) *Client {
	conf := ClientConfig()

	// TODO: add custom transport and other http client configurations

	return &Client{
		StartTime: time.Now().UTC(),
		Conf:      conf,
		User:      user,
		UserID:    userID,
		SfDir:     filepath.Join(conf.Root, "state"),
		Db:        db.NewQuery(filepath.Join(conf.Root, "dbs"), true),
		client: &http.Client{
			Timeout: time.Second * 30,
		},
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
		if err := os.Remove(entry.Name()); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) SaveState() error {
	if err := c.cleanSfDir(); err != nil {
		return err
	}
	// write out
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode json: %v", err)
	}
	fn := fmt.Sprintf("client-state-%s.json", time.Now().UTC().Format("2006-01-02T15-04-05Z"))
	fp := filepath.Join(c.SfDir, fn)
	return os.WriteFile(fp, data, svc.PERMS)
}
