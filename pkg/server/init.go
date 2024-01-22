package server

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/db"
	"github.com/sfs/pkg/env"
	svc "github.com/sfs/pkg/service"
)

// ------- init ---------------------------------------

// create a new SFS service with databases.
// requires a configured .env file prior to running.
func Build() error {
	_, err := SvcInit(svcCfg.SvcRoot)
	if err != nil {
		return fmt.Errorf("failed to initialize new SFS service: %v", err)
	}
	return nil
}

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
		svc, err := SvcInit(svcCfg.SvcRoot)
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
	svc.InitTime = time.Now().UTC()
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
			root, err := svc.Db.GetDirectory(drive.RootID)
			if err != nil {
				return svc, fmt.Errorf("failed to get drive root: %v", err)
			}
			if root == nil {
				return svc, fmt.Errorf("drive root directory not found")
			}
			drive.Root = svc.Populate(root)
			svc.Drives[drive.ID] = drive
		}
	}
	// update state file so we can make each
	// successive service loads quicker. hopefully.
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
func SvcInit(svcRoot string) (*Service, error) {
	// make root service directory (wherever it should located)
	log.Print("[INFO] creating root service directory...")
	if err := os.Mkdir(svcRoot, 0644); err != nil {
		return nil, fmt.Errorf("failed to make service root directory: %v", err)
	}

	//create top-level service directories
	log.Print("[INFO] creating service subdirectories...")
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
	log.Print("[INFO] creating service databases...")
	if err := db.InitDBs(svcPaths[2]); err != nil {
		return nil, fmt.Errorf("failed to initialize service databases: %v", err)
	}

	// create new service instance and save initial state
	log.Print("[INFO] initializng new service instance...")
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

	log.Print("[INFO] all set :)")
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
	// ensure (at least) the necessary dbs and state files are present
	sfPath, err := preChecks(svcPath)
	if err != nil {
		return nil, err
	}
	// unmarshal json state file to a service instance
	svc, err := svcLoad(sfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize service: %v", err)
	}
	// attempt to populate from users and drive databases if state file had no user
	// or drive data.
	if len(svc.Users) == 0 {
		_, err := loadUsers(svc)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve user data: %v", err)
		}
	}
	if len(svc.Drives) == 0 {
		_, err := loadDrives(svc)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve drive data: %v", err)
		}
	} else {
		// load each drive
		for _, d := range svc.Drives {
			root, err := svc.Db.GetDirectory(d.RootID)
			if err != nil {
				return nil, fmt.Errorf("failed to get root directory: %v", err)
			}
			if root == nil {
				return nil, fmt.Errorf("no root directory found for drive: %s", d.ID)
			}
			d.Root = svc.Populate(root)
		}
	}
	return svc, nil
}

// re-initialize users map and service state file
func (s *Service) resetUserMap() {
	s.Users = nil
	s.Users = make(map[string]*auth.User)
}

func (s *Service) resetDrivesMap() {
	s.Drives = nil
	s.Drives = make(map[string]*svc.Drive)
}

// SFS reset function to clear dbs (not delete tables, just all data),
// reset sfs state file, and user root directory. primarily for testing.
func (s *Service) Reset(svcPath string) error {
	if s.AdminMode {
		// clear all databases
		for _, dbName := range s.Db.DBs {
			if err := s.Db.ClearTable(dbName); err != nil {
				return fmt.Errorf("failed to clear databases: %v", err)
			}
		}
		// clear users directory of all roots and subdirectories
		for _, user := range s.Users {
			if err := Clean(user.SvcRoot); err != nil {
				return fmt.Errorf("failed to remove user files: %v", err)
			}
		}
		// reset internal data structures
		s.resetUserMap()
		s.resetDrivesMap()
		if err := s.SaveState(); err != nil {
			return fmt.Errorf("failed to save state file: %v", err)
		}
	} else {
		log.Print("[INFO] must be in admin mode to reset service")
	}
	return nil
}
