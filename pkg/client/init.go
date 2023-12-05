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
	"github.com/sfs/pkg/monitor"
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
|   |   |---users
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
func setup(svcRoot string, e *env.Env) (*Client, error) {
	clientName, err := e.Get("CLIENT")
	if err != nil {
		return nil, err
	}

	// make client service root directory
	svcDir := filepath.Join(svcRoot, clientName)
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
	if err := db.InitDBs(svcPaths[0]); err != nil {
		return nil, err
	}

	// set up new user and initialize a new drive
	newUser, err := newUser(clientName, svcDir, e)
	if err != nil {
		return nil, err
	}

	// initialize a new client for the new user
	client := NewClient(newUser)
	newUser.DriveID = client.Drive.ID

	// save user and drive to db
	if err := client.Db.AddUser(newUser); err != nil {
		return nil, err
	}
	if err := client.Db.AddDrive(client.Drive); err != nil {
		return nil, err
	}
	if err := client.Db.AddDir(client.Drive.Root); err != nil {
		return nil, err
	}

	// set .env file CLIENT_NEW_SERVICE to false so we don't reinitialize every time
	if err := e.Set("CLIENT_NEW_SERVICE", "false"); err != nil {
		return nil, err
	}

	return client, nil
}

// these pull user info from a .env file for now.
// will probably eventually need a way to input an actual new user from a UI
func newUser(clientName string, drvRoot string, e *env.Env) (*auth.User, error) {
	userName, err := e.Get("CLIENT_USERNAME")
	if err != nil {
		return nil, err
	}
	userEmail, err := e.Get("CLIENT_EMAIL")
	if err != nil {
		return nil, err
	}
	newUser := auth.NewUser(clientName, userName, userEmail, drvRoot, false)
	if err := e.Set("CLIENT_ID", newUser.ID); err != nil {
		return nil, err
	}
	return newUser, nil
}

// initial client service set up
func Setup(e *env.Env) (*Client, error) {
	c := ClientConfig()
	client, err := setup(c.Root, e)
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
func LoadClient(usersName string) (*Client, error) {
	// load client state
	data, err := loadStateFile(usersName)
	if err != nil {
		return nil, err
	}
	client := new(Client)
	if err := json.Unmarshal(data, client); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state file: %v", err)
	}

	// load user
	user, err := client.GetUser()
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	// load drive
	drive, err := client.Db.GetDrive(user.DriveID)
	if err != nil {
		return nil, err
	}

	// get root directory for the drive and create a sync index if necessary
	root, err := client.Db.GetDirectory(drive.RootID)
	if err != nil {
		return nil, err
	}
	if root == nil {
		return nil, fmt.Errorf("no root directory for drive %v", drive.ID)
	}
	drive.Root = root
	if drive.SyncIndex == nil {
		drive.SyncIndex = drive.Root.WalkS(svc.NewSyncIndex(drive.OwnerID))
	}
	client.Drive = drive
	client.Root = drive.Root.Path

	// add transfer component
	client.Transfer = transfer.NewTransfer(client.Conf.Port)

	// start client file monitoring services
	client.Monitor = monitor.NewMonitor(client.Root)
	if err := client.StartMonitor(); err != nil {
		return nil, fmt.Errorf("failed to start client monitor: %v", err)
	}

	// initialize handlers map and start all handlers
	client.Handlers = make(map[string]EHandler)
	client.BuildHandlers()
	if err := client.StartHandlers(); err != nil {
		return nil, fmt.Errorf("failed to start event handlers: %v", err)
	}

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
