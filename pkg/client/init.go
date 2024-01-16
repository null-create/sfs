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
|   |   |---client-state-d-m-y-hh-mm-ss.json
|   |---dbs/
|   |   |---users
|   |   |---files
|   |   |---directories
|   |   |---drives

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
func Setup() (*Client, error) {
	// get environment variables and client envCfg
	envCfg := env.NewE()

	// make client service root directory
	svcDir := filepath.Join(cfgs.Root, cfgs.User)
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

	// set up new user
	newUser, err := newUser(svcDir)
	if err != nil {
		return nil, err
	}

	// initialize a new client for the new user
	client, err := NewClient(newUser)
	if err != nil {
		return nil, err
	}
	client.User = newUser
	newUser.DriveID = client.Drive.ID
	newUser.DrvRoot = client.Drive.Root.Path

	// save user, user's root, and drive to db
	if err := client.Db.AddUser(newUser); err != nil {
		return nil, err
	}
	if err := client.Db.AddDir(client.Drive.Root); err != nil {
		return nil, err
	}
	if err := client.Db.AddDrive(client.Drive); err != nil {
		return nil, err
	}

	// set .env file CLIENT_NEW_SERVICE to false so we don't reinitialize every time
	if err := envCfg.Set("CLIENT_NEW_SERVICE", "false"); err != nil {
		return nil, err
	}

	return client, nil
}

// pulls user info from a .env file for now.
// will probably eventually need a way to input an actual new user from a UI
func newUser(drvRoot string) (*auth.User, error) {
	envCfg := env.NewE()
	newUser := auth.NewUser(
		cfgs.User,
		cfgs.UserAlias,
		cfgs.Email,
		cfgs.Root,
		cfgs.IsAdmin,
	)
	if err := envCfg.Set("CLIENT_ID", newUser.ID); err != nil {
		return nil, err
	}
	return newUser, nil
}

func loadStateFile() ([]byte, error) {
	sfDir := filepath.Join(cfgs.Root, cfgs.User, "state")
	entries, err := os.ReadDir(sfDir)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("no state file found for client")
	} else if len(entries) > 1 {
		log.Printf("[WARNING] more than one state file found for client! will default to most recent entry")
		for i, entry := range entries {
			log.Printf("	%d: %s", i+1, entry.Name())
		}
	}
	// get most recent one (assuming more than one present somehow)
	sf := entries[len(entries)-1]
	data, err := os.ReadFile(filepath.Join(sfDir, sf.Name()))
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %v", err)
	}
	return data, nil
}

// loads and populates the users drive and root directory tree.
func loadDrive(client *Client) error {
	drive, err := client.Db.GetDrive(client.Drive.ID)
	if err != nil {
		return err
	}
	if drive == nil {
		return fmt.Errorf("no drive found for user (id=%v)", client.UserID)
	}
	root, err := client.Db.GetDirectory(drive.RootID)
	if err != nil {
		return err
	}
	if root == nil {
		return fmt.Errorf("no root directory for drive (id=%v)", drive.ID)
	}
	drive.Root = client.Populate(root)
	client.Drive = drive
	client.Drive.IsLoaded = true
	client.Root = drive.Root.Path
	return nil
}

// load client from state file, if possible.
// does not start client services. use client.Start()
// to start monitoring and synchronization services.
func LoadClient() (*Client, error) {
	// load client state
	data, err := loadStateFile()
	if err != nil {
		return nil, err
	}
	client := new(Client)
	if err := json.Unmarshal(data, &client); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state file: %v", err)
	}

	// load user if necessary
	if client.User == nil {
		if client.UserID == "" {
			return nil, fmt.Errorf("missing user id")
		}
		user, err := client.Db.GetUser(client.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %v", err)
		} else if user == nil {
			return nil, fmt.Errorf("user %s not found", client.UserID)
		}
		client.User = user
	}

	// load drive with users directory tree.
	if err := loadDrive(client); err != nil {
		return nil, fmt.Errorf("failed to load drive: %v", err)
	}

	// create (or refresh) sync index
	client.Drive.SyncIndex = svc.BuildSyncIndex(client.Drive.Root)

	// add token validation and generation componet
	client.Tok = auth.NewT()

	// set up server endpoints map
	client.setEndpoints()

	// add transfer component
	client.Transfer = transfer.NewTransfer(client.Conf.Port)

	// add monitoring component
	client.Monitor = monitor.NewMonitor(client.Drive.Root.Path)

	// initialize handlers map
	if err := client.BuildHandlers(); err != nil {
		return nil, fmt.Errorf("failed to initialize handlers: %v", err)
	}

	// TODO: pull sync index from server and compare against local index,
	// then make changes as necessary. this should be part of the standard
	// start up process for LoadClient()

	client.StartTime = time.Now().UTC()
	return client, nil
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

// initialize client service
func Init(newClient bool) (*Client, error) {
	if newClient {
		client, err := Setup()
		if err != nil {
			return nil, err
		}
		return client, nil
	} else {
		client, err := LoadClient()
		if err != nil {
			return nil, err
		}
		return client, nil
	}
}
