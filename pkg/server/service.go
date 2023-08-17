package server

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/files"
)

/*
Service Drive directory.
Container for all user Drives (dirs w/metadata and a sub dir for the users stuff).

Top level entry point for internal user file system and their operations.
Will likely be the entry point used for when a server is spun up.

All service configurations may end up living here.
*/
type Service struct {
	InitTime time.Time `json:"init_time"`

	// Drive directory path for sfs service on the server
	ServiceRoot string `json:"service_root"`

	// file name for the sfs state file
	StateFile string `json:"state_file"`

	// path to the service state file
	StateFileFolder string `json:"state_file_folder"`

	// admin mode. allows for expanded permissions when working with
	// the internal sfs file systems.
	AdminMode bool   `json:"admin_mode"`
	Admin     string `json:"admin"`
	AdminKey  string `json:"admin_key"`

	// key: drive-id, val is user struct.
	//
	// user structs contain a pointer to the users Drive directory,
	// so this can be used for measuring disc size and executing
	// health checks
	Users map[string]*auth.User `json:"users"`
}

func Init(new bool, admin bool) *Service {
	c := GetServiceConfig()

	statFileFolder := filepath.Join(c.ServiceRoot, "state")
	stateFile := filepath.Join(statFileFolder, "state.json")

	svc := &Service{
		InitTime:        time.Now(),
		ServiceRoot:     c.ServiceRoot,
		StateFile:       stateFile,
		StateFileFolder: statFileFolder,
		AdminMode:       admin,
		Users:           make(map[string]*auth.User, 0),
	}

	// input server admin credentials if necessary
	if admin {
		s := SrvConfig()
		svc.AdminMode = true
		svc.Admin = s.Server.Admin
		svc.AdminKey = s.Server.AdminKey
	}
	if !new {
		// load from state file and dbs
		if err := svc.Load(); err != nil {
			log.Fatalf("[ERROR] failed to load state file: %v", err)
		}
	} else {
		// initialize new sfs service
		if err := SvcInit(c.ServiceRoot); err != nil {
			log.Fatalf("[ERROR] service init failed: %v", err)
		}
	}
	return svc
}

/*
initialize a new service and db's

generate a root directory for a new sfs service.
the root sfs service directory should have the following structure:

root/
|---users/
|   |---userDriveA/
|   |---userDriveB/
|   (etc)
|---state/
|   |----sfs-state-date:hour:min:sec.json
*/
func SvcInit(path string) error {
	// make root service directory (wherever it should located)
	if err := os.MkdirAll(path, 0666); err != nil {
		log.Fatalf("[ERROR] failed to make service root directory: %v", err)
	}

	paths := []string{filepath.Join(path, "users"), filepath.Join(path, "state")}

	// create user and state sub directories
	for _, p := range paths {
		if err := os.Mkdir(p, 0666); err != nil {
			log.Fatalf("[ERROR] failed to make directory: %v", err)
		}
	}

	return nil
}

// returns the service run time in seconds
func (s *Service) RunTime() float64 {
	return time.Since(s.InitTime).Seconds()
}

// ------- init ---------------------------------------

// read in an external service state file (json) to
// populate the internal data structures.
//
// populates users map through querying the users database
func (s *Service) Load() error {
	return nil
}

/*
SaveState is meant to capture the current value of the following fields:

	InitTime time.Time `json:"init_time"`

	// Drive directory path for sfs service on the server
	ServiceRoot string `json:"service_root"`

	// admin mode. allows for expanded permissions when working with
	// the internal sfs file systems.
	AdminMode bool   `json:"admin_mode"`
	Admin     string `json:"admin"`
	AdminKey  string `json:"admin_key"`

all information about user file metadata  are saved in the database.
the above fields are saved as a json file.
*/
func (s *Service) SaveState() error {
	return nil
}

// get total size (in kb!) of all active user drives
func (s *Service) TotalSize() float64 {
	if len(s.Users) == 0 {
		log.Printf("[DEBUG] no drives to measure")
		return 0.0
	}
	var total float64
	for _, usr := range s.Users {
		total += usr.Drive.DriveSize()
	}
	return total / 1000
}

// ------- new service set up --------------------------------

// TODO: test!
//
// generate some base line meta data for this service instance.
// should generate a users.json file (which will keep track of active users),
// and a drives.json, containing info about each drive, its total size, its location,
// owner, init date, passwords, etc.
func (s *Service) GenBaseFiles(DrivePath string) {
	// create Drive directory
	if err := os.MkdirAll(DrivePath, 0666); err != nil {
		log.Fatalf("[ERROR] failed to create Drive directory \n%v\n", err)
	}
	fileNames := []string{"user-info.json", "drive-info.json", "credentials.json"}
	for i := 0; i < len(fileNames); i++ {
		save(DrivePath, fileNames[i], make(map[string]interface{}))
	}
}

// Build a new privilaged Drive directory for a client on a Nimbus server
//
// Must be under /root/users/<username>
func (s *Service) AllocateDrive(name string, owner string) *files.Drive {
	drivePath := filepath.Join(s.ServiceRoot, name)
	newID := files.NewUUID()

	newRoot := files.NewRootDirectory("root", owner, filepath.Join(drivePath, "root"))
	drive := files.NewDrive(newID, name, owner, drivePath, newRoot)

	s.GenBaseFiles(drivePath)

	return drive
}

// ------- user methods --------------------------------

func (s *Service) TotalUsers() int {
	return len(s.Users)
}

func (s *Service) GetUser(id string) (*auth.User, error) {
	if usr, ok := s.Users[id]; ok {
		return usr, nil
	} else {
		return nil, fmt.Errorf("[ERROR] user %s not found", id)
	}
}

func (s *Service) GetUsers() map[string]*auth.User {
	if len(s.Users) == 0 {
		log.Printf("[DEBUG] no users found")
		return nil
	}
	return s.Users
}

func (s *Service) AddUser(u *auth.User) {
	if _, ok := s.Users[u.ID]; !ok {
		s.Users[u.ID] = u
	} else {
		log.Printf("[DEBUG] user (id=%s) already present", u.ID)
	}
}

func (s *Service) RemoveUser(id string) error {
	if usr, ok := s.Users[id]; ok {
		// remove all user directory and file contents if necessary
		if len(usr.Drive.Root.Dirs) != 0 {
			if err := usr.Drive.Root.Clean(usr.Drive.Root.RootPath); err != nil {
				return fmt.Errorf("[ERROR] unable to remove user and drive contents: %v", err)
			}
		}
		// remove from User directory map
		delete(s.Users, usr.Drive.ID)
	} else {
		return fmt.Errorf("[ERROR] user (id=%s) not found", id)
	}
	return nil
}

// clear all active users drives and deletes all content within
func (s *Service) ClearAll(adminKey string) {
	if s.AdminMode {
		if adminKey == s.AdminKey {
			if len(s.Users) == 0 {
				log.Printf("[DEBUG] no drives to remove")
				return
			}
			// remove all files and directories for this user
			log.Print("[DEBUG] cleaning...")
			for _, usr := range s.Users {
				usr.Drive.Root.Clean(usr.Drive.Root.RootPath)
				delete(s.Users, usr.Drive.ID)
				log.Printf("[DEBUG] user %s was removed", usr.UserName)
			}
			log.Print("[DEBUG] ...done")
		} else {
			log.Print("[DEBUG] enter admin password to clear all user drives")
		}
	} else {
		log.Print("[DEBUG] must be in admin mode to run s.ClearAll()")
	}
}
