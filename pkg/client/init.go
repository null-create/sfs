package client

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/configs"
	"github.com/sfs/pkg/db"
	"github.com/sfs/pkg/logger"
	"github.com/sfs/pkg/monitor"
	svc "github.com/sfs/pkg/service"
	"github.com/sfs/pkg/transfer"
)

/*
define service directory paths, create necessary state file and
database directories, and create service databases and initial state file

root/
|		user/
|   |---backups/      <------ users files and directories backup location
|	  |---root/         <------ users files and directories live here
|	  |---state/
|   |   |---client-state-d-m-y-hh-mm-ss.json
|   |---recycle/      <------ "deleted" files and directories live here.
|   |---dbs/          <------ client side service databases
|   |   |---users
|   |   |---files
|   |   |---directories
|   |   |---drives

users files and directories within a dedicated service root.
"root" here means a dedicated directory for the user to backup and retrieve
any files and directories they wish.

this can allow for more individual control over files and directories
as well as elmininate the need for a dedicated "root" service directory.
(not that this is an inherently bad idea, just want flexiblity)
*/
func SetupClient(svcRoot string) (*Client, error) {
	var setupLog = logger.NewLogger("CLIENT_SETUP", "None")

	// get service configs
	sfcCfg := configs.NewSvcConfig()

	// make client service root directory
	setupLog.Info("making SFS service directories...")
	svcDir := filepath.Join(svcRoot, cCfgs.User)
	if err := os.Mkdir(svcDir, svc.PERMS); err != nil {
		return nil, err
	}

	// define service directory paths & make directories
	svcPaths := []string{
		filepath.Join(svcDir, "dbs"),
		filepath.Join(svcDir, "root"),
		filepath.Join(svcDir, "state"),
		filepath.Join(svcDir, "recycle"),
		filepath.Join(svcDir, "backups", "root"),
	}
	for _, dirPath := range svcPaths {
		if err := os.MkdirAll(dirPath, svc.PERMS); err != nil {
			return nil, err
		}
	}

	// make each database
	setupLog.Info("creating databases...")
	if err := db.InitServerDBs(svcPaths[0]); err != nil {
		return nil, err
	}

	// set up new user
	setupLog.Info("creating user...")
	newUser, err := newUser()
	if err != nil {
		return nil, err
	}

	// initialize a new client for the new user
	setupLog.Info("creating client...")
	client, err := NewClient(newUser)
	if err != nil {
		return nil, err
	}
	newUser.DriveID = client.Drive.ID
	newUser.DrvRoot = client.Drive.Root.ClientPath
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
	if err := sfcCfg.Set(configs.CLIENT_NEW_SERVICE, "false"); err != nil {
		return nil, err
	}

	setupLog.Info("all set :)")
	return client, nil
}

// initialization logger
var initLog = logger.NewLogger("CLIENT_INIT", "None")

// pulls user info from a .env file for now.
// will probably eventually need a way to input an actual new user from a UI
func newUser() (*auth.User, error) {
	newUser := auth.NewUser(
		cCfgs.User,
		cCfgs.UserAlias,
		cCfgs.Email,
		cCfgs.Root,
		cCfgs.IsAdmin,
	)
	if err := svcCfgs.Set(configs.CLIENT_ID, newUser.ID); err != nil {
		initLog.Error("failed to set user ID as an env variable: " + err.Error())
		return nil, err
	}
	return newUser, nil
}

// initialize a new http.Client object
func newHttpClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   1 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}

func loadStateFile() ([]byte, error) {
	sfDir := filepath.Join(cCfgs.Root, cCfgs.User, "state")
	entries, err := os.ReadDir(sfDir)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("no state file found for client")
	} else if len(entries) > 1 {
		var output string
		for i, entry := range entries {
			output += fmt.Sprintf("	%d: %s\n", i+1, entry.Name())
		}
		initLog.Warn("more than one state file found in: " + sfDir + "\n" + output)
	}
	// get most recent one (assuming more than one present somehow)
	sf := entries[len(entries)-1]
	data, err := os.ReadFile(filepath.Join(sfDir, sf.Name()))
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %v", err)
	}
	return data, nil
}

// load client from state file and initialize.
// does not automatically start persistent client services.
//
// set persist to true when using as a long-running operation
// and to utilize real-time monitoring and synchronization services,
// then follow with a call to client.Start() to start a blocking process.
// otherwise set to false if the client should be used for
// one-off operations.
func LoadClient(persist bool) (*Client, error) {
	data, err := loadStateFile()
	if err != nil {
		return nil, err
	}
	client := new(Client)
	if err := json.Unmarshal(data, &client); err != nil {
		initLog.Log(logger.ERROR, fmt.Sprintf("failed to unmarshal state file: %v", err))
		return nil, fmt.Errorf("failed to unmarshal state file: %v", err)
	}

	// initialize logger
	client.log = logger.NewLogger("Client", client.UserID)

	// initialize DB connection
	client.Db = db.NewQuery(client.Db.DBPath, true)

	// load user info
	if err := client.LoadUser(); err != nil {
		initLog.Log(logger.ERROR, fmt.Sprintf("failed to load user: %v", err))
		return nil, fmt.Errorf("failed to load user: %v", err)
	}

	// load drive with users sfs directory tree populated.
	// also refreshes (or generates) drive sync index.
	if err := client.LoadDrive(); err != nil {
		initLog.Log(logger.ERROR, fmt.Sprintf("failed to load drive: %v", err))
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

	// initialize event map
	client.InitHandlerMap()

	// load and start persistent services only when necessary.
	// persist should only be set to true when followed by a
	// call to client.Start(), otherwise none of the monitoring
	// services will be able to actually run.
	if persist {
		// make sure the client was registered registered before starting,
		// if synchronization with the server is enabled. If so, make sure
		// any local items that have been added since our last sync are also
		// registered with the server.
		if client.SvrSync() {
			// register client with server if necessary
			if err := client.RegisterClient(); err != nil {
				initLog.Log(logger.ERROR, fmt.Sprintf("failed to register client: %v", err))
			}

			// register any files with the server that weren't previously registered
			if err := client.RegisterItems(); err != nil {
				initLog.Log(logger.ERROR, fmt.Sprintf("failed to register local items: %v", err))
			}
		}

		// initialize handlers map
		if err := client.BuildHandlers(); err != nil {
			initLog.Log(logger.ERROR, fmt.Sprintf("failed to initialize handlers: %v", err))
			return nil, fmt.Errorf("failed to initialize handlers: %v", err)
		}

		// start monitoring services in SFS root directory
		// creates event handlers for each item thats being monitored.
		if err := client.StartMonitor(); err != nil {
			initLog.Log(logger.ERROR, fmt.Sprintf("failed to start monitoring services: %v", err))
			return nil, fmt.Errorf("failed to start monitoring services: %v", err)
		} else {
			client.log.Info(fmt.Sprintf("monitor is running. watching %d local item(s)", len(client.Monitor.Events)))
		}

		// parse html templates for the web ui
		if err := client.ParseTemplates(); err != nil {
			initLog.Log(logger.ERROR, fmt.Sprintf("failed to parse templates: %v", err))
			return nil, fmt.Errorf("failed to parse templates: %v", err)
		}
	}

	client.StartTime = time.Now().UTC()
	if err := client.SaveState(); err != nil {
		client.log.Error("failed to save state file: " + err.Error())
	}

	initLog.Log(logger.INFO, fmt.Sprintf("client started at: %v", time.Now().UTC()))
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
	c.Endpoints["file info"] = EndpointRootWithPort + "/v1/files/i/" // NOTE: this will need to be concatenated with a file ID
	c.Endpoints["all dirs"] = EndpointRootWithPort + "/v1/dirs/i/all/" + c.UserID
	c.Endpoints["dir info"] = EndpointRootWithPort + "/v1/dirs/i/" // NOTE: this will need to be concatenated with a directory ID
	c.Endpoints["new dir"] = EndpointRootWithPort + "/v1/dirs/new"
	c.Endpoints["drive"] = EndpointRootWithPort + "/v1/drive/" + c.DriveID
	c.Endpoints["new drive"] = EndpointRootWithPort + "/v1/drive/new"
	c.Endpoints["sync"] = EndpointRootWithPort + "/v1/sync/" + c.DriveID
	c.Endpoints["get index"] = EndpointRootWithPort + "/v1/sync/" + c.DriveID
	c.Endpoints["gen index"] = EndpointRootWithPort + "/v1/sync/" + c.DriveID + "/index"
	c.Endpoints["gen updates"] = EndpointRootWithPort + "/v1/sync/" + c.DriveID + "/update"
	c.Endpoints["user"] = EndpointRootWithPort + "/v1/users/" + c.UserID
	c.Endpoints["new user"] = EndpointRootWithPort + "/v1/users/new"
	c.Endpoints["all users"] = EndpointRootWithPort + "/v1/users/all"
	c.Endpoints["runtime"] = EndpointRootWithPort + "/v1/runtime"
}

// creates a new client object. does not create actual service directories or
// other necessary infrastructure -- only the client itself.
func NewClient(user *auth.User) (*Client, error) {
	ccfg := GetClientConfigs()

	// set up local client services
	driveID := auth.NewUUID()
	svcRoot := filepath.Join(ccfg.Root, user.Name)
	root := svc.NewRootDirectory("root", user.ID, driveID, filepath.Join(svcRoot, "root"))
	root.BackupPath = filepath.Join(svcRoot, "backups")
	drv := svc.NewDrive(driveID, user.Name, user.ID, root.ClientPath, root.ID, root)
	drv.Root = root
	drv.IsLoaded = true
	user.DriveID = driveID
	user.DrvRoot = drv.RootPath
	user.SvcRoot = root.ClientPath

	// intialize client
	client := &Client{
		StartTime:      time.Now().UTC(),
		Conf:           ccfg,
		UserID:         user.ID,
		User:           user,
		Root:           filepath.Join(svcRoot, "root"),
		SfDir:          filepath.Join(svcRoot, "state"),
		RecycleBin:     filepath.Join(svcRoot, "recycle"),
		LocalBackupDir: filepath.Join(root.ClientPath, "backups"),
		Endpoints:      make(map[string]string),
		Monitor:        monitor.NewMonitor(drv.Root.ClientPath),
		DriveID:        driveID,
		Drive:          drv,
		Db:             db.NewQuery(filepath.Join(svcRoot, "dbs"), true),
		log:            logger.NewLogger("Client", user.ID),
		Tok:            auth.NewT(),
		Handlers:       make(map[string]Handler),
		Transfer:       transfer.NewTransfer(),
		Client:         newHttpClient(),
	}

	// add drive itself to DB (root was added during discovery)
	// then attach to client
	if err := client.Db.AddDrive(drv); err != nil {
		return nil, fmt.Errorf("failed to add client drive to database: %v", err)
	}
	if err := client.Db.AddDir(root); err != nil {
		return nil, fmt.Errorf("failed to add root directory to database: %v", err)
	}
	client.Drive = drv

	// build services endpoints map (files and directories have endpoints defined
	// within their respective data structures)
	client.setEndpoints()

	// add token component
	client.Tok = auth.NewT()

	// initialize local sync index
	client.BuildSyncIndex()

	// register drive with the server if autosync is enabled, if not defaulting
	// to using local storage.
	if client.SvrSync() {
		if err := client.RegisterClient(); err != nil {
			client.log.Warn("failed to register client with server: " + err.Error())
		}
	}
	// save initial state
	if err := client.SaveState(); err != nil {
		return nil, fmt.Errorf("failed to save initial state: %v", err)
	}
	return client, nil
}

// initialize client service
func Init(newClient bool) (*Client, error) {
	if newClient {
		client, err := SetupClient(cCfgs.Root)
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
