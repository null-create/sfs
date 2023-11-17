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
	"github.com/sfs/pkg/env"
	svc "github.com/sfs/pkg/service"
	"github.com/sfs/pkg/transfer"
)

/*
define service directory paths, create necessary state file and
database directories, and create service databases and initial state file

root/
|		user
|	  |---root/     <------ users files and directories live here
|	  |---state/
|   |   |---user-state-date-time.json
|   |---dbs/
|   |   |---files
|   |   |---directories
users files and directories within a dedicated service root.
"root" here means a dedicated directory for the user to backup and retrieve
any files and directories they wish.

NOTE: A future alternative mode may allow for individual files spread
across a user's normal system to be "marked" as files to "watch" for
activity (such as updates, modifications, etc), and then be queued for
synching or backing up with the server (automatically, or when a user
manually intiates a sync).

this can allow for more individual control over files and directories
as well as elmininate the need for a dedicated "root" service directory.
(not that this is an inherently bad idea, just want flexiblity)
*/
func setup(user, svcRoot string, e *env.Env) (*Client, error) {
	// make client service root directory
	svcDir := filepath.Join(svcRoot, user)
	if err := os.Mkdir(svcDir, svc.PERMS); err != nil {
		return nil, err
	}

	// define service directory paths & make directories
	svcPaths := []string{
		filepath.Join(svcDir, "dbs"),
		filepath.Join(svcDir, "root"),
		filepath.Join(svcDir, "state"),
	}
	for _, svcPath := range svcPaths {
		if err := os.Mkdir(svcPath, svc.PERMS); err != nil {
			return nil, err
		}
	}

	// make each database
	if err := db.InitClientDBs(svcPaths[0]); err != nil {
		return nil, err
	}

	// initialize a new client with a new user
	client, err := newClient(user)
	if err != nil {
		return nil, err
	}

	// set .env file CLIENT_NEW_SERVICE to false so we don't reinitialize every time
	if err := e.Set("CLIENT_NEW_SERVICE", "false"); err != nil {
		return nil, err
	}

	return client, nil
}

func newClient(user string) (*Client, error) {
	client := NewClient(user, auth.NewUUID())
	if err := client.SaveState(); err != nil {
		return nil, err
	}
	return client, nil
}

// these pull user info from a .env file for now.
// will probably eventually need a way to input an actual new user from a UI
func newUser(user string, driveID string, drvRoot string, e *env.Env) (*auth.User, error) {
	userName, err := e.Get("CLIENT_USERNAME")
	if err != nil {
		return nil, err
	}
	userEmail, err := e.Get("CLIENT_EMAIL")
	if err != nil {
		return nil, err
	}
	newUser := auth.NewUser(user, userName, userEmail, driveID, drvRoot, false)
	return newUser, nil
}

// initial client service set up
func Setup(e *env.Env) (*Client, error) {
	c := ClientConfig()
	client, err := setup(c.User, c.Root, e)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func loadStateFile(user string) ([]byte, error) {
	c := ClientConfig()
	fp := filepath.Join(c.Root, user, "state")
	entries, err := os.ReadDir(fp)
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
	// get most recent one (assuming more than one present somehow)
	sf := entries[len(entries)-1]
	data, err := os.ReadFile(filepath.Join(fp, sf.Name()))
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %v", err)
	}
	return data, nil
}

// load client from state file, if possible
func LoadClient(user string) (*Client, error) {
	data, err := loadStateFile(user)
	if err != nil {
		return nil, err
	}
	client := new(Client)
	if err := json.Unmarshal(data, client); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state file: %v", err)
	}
	// TODO: transfer component is nil when loaded from a state file.
	// will need a way to instantiate here
	client.Transfer = transfer.NewTransfer()

	// start client file monitoring services
	if err := client.StartMonitor(); err != nil {
		return nil, fmt.Errorf("failed to start client monitor: %v", err)
	}
	// initialize handlers for all monitoring goroutines
	client.BuildHandlers()

	client.StartTime = time.Now().UTC()

	return client, nil
}

// initialize client service
func Init(newClient bool) (*Client, error) {
	e := env.NewE()
	if newClient {
		client, err := Setup(e)
		if err != nil {
			return nil, err
		}
		return client, nil
	} else {
		user, err := e.Get("CLIENT")
		if err != nil {
			return nil, err
		}
		client, err := LoadClient(user)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
}
