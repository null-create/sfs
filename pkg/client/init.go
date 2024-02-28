package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
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
|   |---recycle/   <------ "deleted" files and directories live here.
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
	log.Printf("[INFO] making SFS service directories...")
	svcDir := filepath.Join(cfgs.Root, cfgs.User)
	if err := os.Mkdir(svcDir, svc.PERMS); err != nil {
		return nil, err
	}

	// define service directory paths & make directories
	svcPaths := []string{
		filepath.Join(svcDir, "dbs"),
		filepath.Join(svcDir, "root"),
		filepath.Join(svcDir, "state"),
		filepath.Join(svcDir, "recycle"),
	}
	for _, dirPath := range svcPaths {
		if err := os.Mkdir(dirPath, svc.PERMS); err != nil {
			return nil, err
		}
	}

	// make each database
	log.Printf("[INFO] creating databases...")
	if err := db.InitDBs(svcPaths[0]); err != nil {
		return nil, err
	}

	// set up new user
	log.Printf("[INFO] creating user...")
	newUser, err := newUser()
	if err != nil {
		return nil, err
	}

	// initialize a new client for the new user
	log.Printf("[INFO] creating client...")
	client, err := NewClient(newUser)
	if err != nil {
		return nil, err
	}
	newUser.DriveID = client.Drive.ID
	newUser.DrvRoot = client.Drive.Root.Path
	client.User = newUser
	client.UserID = newUser.ID

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
	// attempt to register drive with server.
	if client.autoSync() {
		if err := client.RegisterDrive(); err != nil {
			return nil, err
		}
	}
	// set .env file CLIENT_NEW_SERVICE to false so we don't reinitialize every time
	if err := envCfg.Set("CLIENT_NEW_SERVICE", "false"); err != nil {
		return nil, err
	}

	log.Printf("[INFO] all set :)")
	return client, nil
}

// pulls user info from a .env file for now.
// will probably eventually need a way to input an actual new user from a UI
func newUser() (*auth.User, error) {
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

// initialize a new http.Client object
func newHttpClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment, // TODO add a proxy configuration?
			Dial: (&net.Dialer{
				Timeout:   1 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
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

// load client from state file, if possible.
// does not start client services. use client.Start()
// to start monitoring and synchronization services.
//
// set persist to true when using as a long-running operation
// and to utilize real-time monitoring and synchronization services,
// otherwise set to false if the client should be used for
// one-off operations.
func LoadClient(persist bool) (*Client, error) {
	// load client state
	data, err := loadStateFile()
	if err != nil {
		return nil, err
	}
	client := new(Client)
	if err := json.Unmarshal(data, &client); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state file: %v", err)
	}

	// load user info
	if err := client.LoadUser(); err != nil {
		return nil, fmt.Errorf("failed to load user: %v", err)
	}

	// load drive with users sfs directory tree populated.
	// also refreshes (or generates) drive sync index.
	if err := client.LoadDrive(); err != nil {
		return nil, fmt.Errorf("failed to load drive: %v", err)
	}

	// initialize http client
	client.Client = newHttpClient()

	// add token validation and generation component
	client.Tok = auth.NewT()

	// set up server endpoints map
	client.setEndpoints()

	// add transfer component
	client.Transfer = transfer.NewTransfer()

	// add monitoring component
	client.Monitor = monitor.NewMonitor(client.Root)

	// initialize event maps
	client.InitHandlerMaps()

	// load and start persistent services only when necessary.
	// persist should only be set to true when followed by a
	// call to client.Start(), otherwise none of the monitoring
	// services will be able to actually run.
	if persist {
		// make sure the drive was registered before starting
		if !client.Drive.IsRegistered() {
			if err := client.RegisterDrive(); err != nil {
				return nil, err
			}
		}
		// start monitoring services in SFS root directory
		if err := client.StartMonitor(); err != nil {
			return nil, fmt.Errorf("failed to start monitoring services: %v", err)
		}
		// initialize handlers map
		if err := client.BuildHandlers(); err != nil {
			return nil, fmt.Errorf("failed to initialize handlers: %v", err)
		}
		// start event handlers
		if err := client.StartHandlers(); err != nil {
			return nil, fmt.Errorf("failed to start event handlers: %v", err)
		}

		// TODO: pull sync index from server and compare against local index,
		// then make changes as necessary. this should be part of the standard
		// start up process for LoadClient()
		// if c.autoSync() {

		// }

		client.StartTime = time.Now().UTC()
	}
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
	root := svc.NewRootDirectory("root", ccfg.UserID, driveID, filepath.Join(svcRoot, "root"))
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
		RecycleBin:  filepath.Join(svcRoot, "recycle"),
		Endpoints:   make(map[string]string),
		Monitor:     monitor.NewMonitor(drv.Root.Path),
		DriveID:     driveID,
		Drive:       drv,
		Db:          db.NewQuery(filepath.Join(svcRoot, "dbs"), true),
		Handlers:    make(map[string]func()),
		OffSwitches: make(map[string]chan bool),
		Transfer:    transfer.NewTransfer(),
		Client:      newHttpClient(),
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
	// register drive with the server if autosync is enabled
	if c.autoSync() {
		if err := c.RegisterDrive(); err != nil {
			return nil, fmt.Errorf("failed to register drive: %v", err)
		}
	}

	// build services endpoints map (files and directories have endpoints defined
	// within their respective data structures)
	c.setEndpoints()

	// add token component
	c.Tok = auth.NewT()

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
		client, err := LoadClient(true)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
}
