package client

import (
	"encoding/json"
	"fmt"
	"log"
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

	User   string `json:"user"`       // user name
	UserID string `json:"user_id"`    // user ID
	SfPath string `json:"state_file"` // path to state file

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
		SfPath:    filepath.Join(conf.Root, "state"),
		Db:        db.NewQuery(filepath.Join(conf.Root, "dbs"), true),
		client: &http.Client{
			Timeout: time.Second * 30,
		},
	}
}

func loadStateFile() ([]byte, error) {
	c := ClientConfig()

	entries, err := os.ReadDir(filepath.Join(c.Root, "state"))
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		log.Fatal("no state file found for client!")
	} else if len(entries) > 1 {
		log.Printf("[WARNING] more than one state file found for client! will default to most recent entry")
		for i, entry := range entries {
			log.Printf("	%d: %s", i+1, entry.Name())
		}
	}

	sf := entries[len(entries)-1] // get most recent one (assuming more than one present somehow)
	data, err := os.ReadFile(sf.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %v", err)
	}
	return data, nil
}

// load client from state file, if possible
func LoadClient() (*Client, error) {
	data, err := loadStateFile()
	if err != nil {
		return nil, err
	}
	client := new(Client)
	if err := json.Unmarshal(data, client); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state file: %v", err)
	}
	return client, nil
}

func (c *Client) SaveState() error {
	// remove previous state file(s)
	conf := ClientConfig()
	entries, err := os.ReadDir(filepath.Join(conf.Root, "state"))
	if err != nil {
		return nil
	}
	for _, entry := range entries {
		if err := os.Remove(entry.Name()); err != nil {
			return err
		}
	}

	// write out
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode json: %v", err)
	}
	fn := fmt.Sprintf("client-state-%s.json", time.Now().UTC().Format("2006-01-02T15-04-05Z"))
	fp := filepath.Join(c.SfPath, fn)
	return os.WriteFile(fp, data, svc.PERMS)
}
