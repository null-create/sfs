package server

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sfs/pkg/db"
	"github.com/sfs/pkg/env"
	"github.com/sfs/pkg/logger"
	svc "github.com/sfs/pkg/service"
)

// ------- init ---------------------------------------

// initialize a new server-side sfs service from either a state file/dbs or
// create a new service from scratch.
func Init(new bool, admin bool) (*Service, error) {
	if !new {
		// load from state file and dbs
		svc, err := SvcLoad(svcCfg.SvcRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to load service: %v", err)
		}
		if admin {
			setAdmin(svc)
		}
		return svc, nil
	} else {
		// initialize new sfs service
		svc, err := SetUpService(svcCfg.SvcRoot)
		if err != nil {
			return nil, err
		}
		if admin {
			setAdmin(svc)
		}
		return svc, nil
	}
}

func setAdmin(svc *Service) {
	svc.AdminMode = true
	svc.Admin = svrCfg.Admin
	svc.AdminKey = svrCfg.AdminKey
}

// searches for the service state file.
// will returna an empty string if there was an error
func findStateFile(svcRoot string) (string, error) {
	sfPath := filepath.Join(svcRoot, "state")
	entries, err := os.ReadDir(sfPath)
	if err != nil {
		return "", fmt.Errorf("unable to read state file directory: %s \n%v", sfPath, err)
	}
	// should only ever be one state file at a time
	if len(entries) > 1 {
		log.Printf("[WARNING] multiple state files found under: %s", sfPath)
		for i, e := range entries {
			log.Printf("%d.	%s\n", i+1, e.Name())
		}
	}
	for _, entry := range entries {
		if strings.Contains(entry.Name(), "sfs-state") {
			return filepath.Join(sfPath, entry.Name()), nil
		}
	}
	return "", nil
}

// load from a service state file. returns a new empty service struct.
//
// *does not instatiate svc, db, or user paths.* must be set elsewhere
func loadStateFile(sfPath string) (*Service, error) {
	file, err := os.ReadFile(sfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %v", err)
	}
	svc := new(Service)
	if err := json.Unmarshal(file, svc); err != nil {
		return nil, fmt.Errorf("failed unmarshal service state file: %v", err)
	}
	svc.StateFile = sfPath
	return svc, nil
}

// populate svc.Users map from users database
func loadUsers(svc *Service) (*Service, error) {
	usrs, err := svc.Db.GetUsers()
	if err != nil {
		return svc, fmt.Errorf("failed to retrieve user data from Users database: %v", err)
	}
	if len(usrs) == 0 {
		return svc, nil
	}
	// NOTE: this assumes the physical files for each
	// user being loaded are already present. calling svc.AddUser()
	// will allocate a new drive service with new base files.
	for _, u := range usrs {
		if _, exists := svc.Users[u.ID]; !exists {
			svc.Users[u.ID] = u
		}
	}
	return svc, nil
}

// load all users files and directories.
func loadDrive(svc *Service, drv *svc.Drive) error {
	files, err := svc.Db.GetUsersFiles(drv.OwnerID)
	if err != nil {
		return err
	}
	if len(files) > 0 {
		drv.Root.AddFiles(files)
	}
	dirs, err := svc.Db.GetUsersDirectories(drv.OwnerID)
	if err != nil {
		return err
	}
	if len(dirs) > 0 {
		drv.Root.AddSubDirs(dirs)
	}
	return nil
}

// loads each drive from the database.
func loadDrives(svc *Service) (*Service, error) {
	drives, err := svc.Db.GetDrives()
	if err != nil {
		return svc, fmt.Errorf("failed to query database for drives: %v", err)
	}
	if len(drives) == 0 {
		return svc, nil
	}
	for _, drive := range drives {
		if _, exists := svc.Drives[drive.ID]; !exists {
			root, err := svc.Db.GetDirectoryByID(drive.RootID)
			if err != nil {
				return svc, fmt.Errorf("failed to get drive root: %v", err)
			}
			if root == nil {
				return svc, fmt.Errorf("drive root directory not found")
			}
			if err := loadDrive(svc, drive); err != nil {
				return svc, fmt.Errorf("failed to load drive: %v", err)
			}
			drive.BuildSyncIdx()
			svc.Drives[drive.ID] = drive
		}
	}
	if err := svc.SaveState(); err != nil {
		return svc, fmt.Errorf("failed to save state: %v", err)
	}
	return svc, nil
}

/*
generate a root directory for a new sfs service.
the root sfs service directory should have the following structure:

root/
|---users/
|   |---userDriveA/
|   |---userDriveB/
|   (etc)
|---state/
|   |----sfs-state-date:hour:min:sec.json
|---dbs/
|   |---users
|   |---drives
|   |---directories
|   |---files
*/

// initialize a new service and corresponding databases
func SetUpService(svcRoot string) (*Service, error) {
	// create a logger for initialization steps
	initLogger := logger.NewLogger("SERVICE_INIT", "None")

	// make root service directory (wherever it should located)
	initLogger.Info("creating root service directory...")
	if err := os.Mkdir(svcRoot, 0644); err != nil {
		return nil, fmt.Errorf("failed to make service root directory: %v", err)
	}

	//create top-level service directories
	initLogger.Info("creating root service directory...")
	svcPaths := []string{
		filepath.Join(svcRoot, "users"),
		filepath.Join(svcRoot, "state"),
		filepath.Join(svcRoot, "dbs"),
	}
	for _, p := range svcPaths {
		if err := os.Mkdir(p, 0644); err != nil {
			return nil, fmt.Errorf("failed to make service directory: %v", err)
		}
	}

	// create new service databases
	initLogger.Info("creating service databases...")
	if err := db.InitServerDBs(svcPaths[2]); err != nil {
		return nil, fmt.Errorf("failed to initialize service databases: %v", err)
	}

	// create new service instance and save initial state
	initLogger.Info("initializing new service instance...")
	svc := NewService(svcRoot)
	if err := svc.SaveState(); err != nil {
		return nil, fmt.Errorf("failed to save service state: %v", err)
	}

	// update NEW_SERVICE variable so future boot ups won't
	// create a new service every time
	e := env.NewE()
	if err := e.Set("NEW_SERVICE", "false"); err != nil {
		return nil, fmt.Errorf("failed to update service .env file: %v", err)
	}

	initLogger.Info("all set :)")
	return svc, nil
}

//   - are the databases present?
//   - is the statefile present?
//
// if not, raise an error, otherwise returns the
// path to the state file upon success
func preChecks(svcRoot string) (string, error) {
	entries, err := os.ReadDir(svcRoot)
	if err != nil {
		return "", err
	}
	if len(entries) == 0 {
		return "", fmt.Errorf("no statefile, user, or database directories found! \n%v", err)
	}
	for _, entry := range entries {
		if entry.Name() == "state" || entry.Name() == "dbs" {
			if isEmpty(filepath.Join(svcRoot, entry.Name())) {
				return "", fmt.Errorf("%s directory is empty \n%v", entry.Name(), err)
			}
		}
	}
	sfPath, err := findStateFile(svcRoot)
	if err != nil {
		return "", err
	}
	if sfPath == "" {
		return "", fmt.Errorf("no state file found")
	}
	return sfPath, nil
}

// intialize sfs service from a state file
func svcLoad(sfPath string) (*Service, error) {
	svc, err := loadStateFile(sfPath)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

// reads in an external service state file and instantiates an SFS service instance.
func SvcLoad(svcPath string) (*Service, error) {
	// logger used during an established service initialization
	var initLogger = logger.NewLogger("SERVICE_INIT", "None")

	// ensure (at least) the necessary dbs and state files are present
	sfPath, err := preChecks(svcPath)
	if err != nil {
		initLogger.Error(err.Error())
		return nil, err
	}
	// unmarshal json state file to a service instance
	svc, err := svcLoad(sfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize service: %v", err)
	}
	// explicity initialize the db connection since its logger
	// doesnt get initialized when we load the service from the
	// state file
	svc.Db = db.NewQuery(svc.DbDir, true)

	// load logger
	svc.log = logger.NewLogger("Service", svc.ID)

	// add configs to service instance
	svc.svcCfgs = svcCfg

	// load users and drives
	_, err = loadUsers(svc)
	if err != nil {
		initLogger.Error(fmt.Sprintf("failed to retrieve user data: %v", err))
		return nil, fmt.Errorf("failed to retrieve user data: %v", err)
	}
	_, err = loadDrives(svc)
	if err != nil {
		initLogger.Error(fmt.Sprintf("failed to retrieve drive data: %v", err))
		return nil, fmt.Errorf("failed to retrieve drive data: %v", err)
	}

	// update state file and start time
	if err := svc.SaveState(); err != nil {
		initLogger.Error(fmt.Sprintf("failed to save service state: %v", err))
		return nil, fmt.Errorf("failed to save state: %v", err)
	}
	svc.InitTime = time.Now().UTC()
	initLogger.Log("INFO", fmt.Sprintf("service loaded at %s", svc.InitTime))
	return svc, nil
}
