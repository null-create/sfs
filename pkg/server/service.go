package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/db"
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
	SvcRoot string `json:"service_root"`

	// path to state file
	StateFile string `json:"sf_file"`

	// path to user file directory
	UserDir string `json:"user_dir"`

	// path to database directory
	DbDir string `json:"db_dir"`

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

// intialize a new empty service struct
func NewService(svcRoot string) *Service {
	return &Service{
		InitTime: time.Now().UTC(),
		SvcRoot:  svcRoot,

		// we don't set StateFile because we
		// assume it doesn't exist when NewService is called
		StateFile: "",

		UserDir: filepath.Join(svcRoot, "users"),
		DbDir:   filepath.Join(svcRoot, "dbs"),
		Users:   make(map[string]*auth.User),
	}
}

// ------- init ---------------------------------------

func setAdmin(svc *Service) {
	cfg := ServerConfig()
	svc.AdminMode = true
	svc.Admin = cfg.Server.Admin
	svc.AdminKey = cfg.Server.AdminKey
}

// determine whether we have a sfs-state-date:hour:min:sec.json file
// under svcroot/state
func hasStateFile(path string) (bool, fs.DirEntry) {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Printf("[ERROR] unable to find %s \n%v\n", path, err)
		return false, nil
	}
	if len(entries) > 1 {
		log.Printf("[WARNING] multiple state files found under %s", path)
		for _, e := range entries {
			log.Printf("	-%s\n", e.Name())
		}
	}
	for _, entry := range entries {
		// NOTE: this might not be the most *current* version,
		// just the one that's present at the moment.
		if strings.Contains(entry.Name(), "sfs-state") {
			return true, entry
		}
	}
	return false, nil
}

// load service state file.
//
// does not instatiate svc, db, or user paths. must be set elsewhere
func loadStateFile(sfPath string) (*Service, error) {
	// load state file and unmarshal into service struct
	file, err := os.ReadFile(sfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %v", err)
	}

	svc := &Service{}
	if err := json.Unmarshal([]byte(file), svc); err != nil {
		return nil, fmt.Errorf("failed unmarshal service state file: %v", err)
	}
	svc.StateFile = sfPath
	svc.InitTime = time.Now().UTC()

	return svc, nil
}

// populate svc.Users map from users database
func loadUsers(svc *Service) (*Service, error) {
	q := db.NewQuery(svc.DbDir)
	usrs, err := q.GetUsers()
	if err != nil {
		return nil, fmt.Errorf("[ERROR] failed to retrieve user data from Users database: %v", err)
	}
	for _, u := range usrs {
		svc.AddUser(u)
	}
	return svc, nil
}

func Init(new bool, admin bool) (*Service, error) {
	c := ServiceConfig()
	if !new {
		// ---- load from state file and dbs
		if ok, entry := hasStateFile(filepath.Join(c.ServiceRoot, "state")); ok {
			svc, err := SvcLoad(entry.Name(), false)
			if err != nil {
				return nil, fmt.Errorf("[ERROR] failed to load service config: %v", err)
			}
			if admin {
				setAdmin(svc)
			}
			return svc, nil
			// no state file found!
		} else {
			return nil, fmt.Errorf("[ERROR] unable to load service config")
		}
	} else {
		// ----- initialize new sfs service
		svc, err := SvcInit(c.ServiceRoot, false)
		if err != nil {
			return nil, fmt.Errorf("[ERROR] %v", err)
		}
		if admin {
			setAdmin(svc)
		}
		return svc, nil
	}
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
|---dbs/
|   |---users
|   |---drives
|   |---directories
|   |---files
*/
func SvcInit(svcPath string, debug bool) (*Service, error) {
	// ------- make root service directory (wherever it should located)
	log.Print("creating root service directory...")
	if err := os.Mkdir(svcPath, 0666); err != nil {
		return nil, fmt.Errorf("[ERROR] failed to make service root directory: %v", err)
	}

	//-------- create top-level service directories
	log.Print("creating service subdirectories...")
	svcPaths := []string{
		filepath.Join(svcPath, "users"),
		filepath.Join(svcPath, "state"),
		filepath.Join(svcPath, "dbs"),
	}
	for _, p := range svcPaths {
		if err := os.Mkdir(p, 0666); err != nil {
			return nil, fmt.Errorf("[ERROR] failed to make service directory: %v", err)
		}
	}

	// -------- create new service databases
	log.Print("creating service databases...")
	if err := db.InitDBs(svcPaths[2]); err != nil {
		return nil, fmt.Errorf("[ERROR] failed to initialize service databases: %v", err)
	}

	// --------- create new service instance and save initial state
	log.Print("initializng new service instance...")
	svc := &Service{
		InitTime:  time.Now().UTC(),
		SvcRoot:   svcPath,
		StateFile: "",
		UserDir:   svcPaths[0],
		DbDir:     svcPaths[2],
		AdminMode: false,
	}
	if err := svc.SaveState(); err != nil {
		return nil, fmt.Errorf("[ERROR] %v", err)
	}
	return svc, nil

}

// read in an external service state file
//
// populates users map through querying the users database
func SvcLoad(sfPath string, debug bool) (*Service, error) {

	// TODO: add some "pre-checks"
	// 	- are the databases present?
	// 	- is the statefile present?

	svc, err := loadStateFile(sfPath)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] %v", err)
	}

	// populate user map via user database
	svc, err = loadUsers(svc)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] %v", err)
	}

	return svc, nil
}

// ------ utils --------------------------------

/*
SaveState is meant to capture the current value of
the following fields when saving service state to disk:

	InitTime time.Time `json:"init_time"`

	SvcRoot string `json:"service_root"`  // directory path for sfs service on the server
	StateFile string `json:"state_file"`  // path to state file directory

	UserDir string `json:"user_dir"` // path to user drives directory
	DbDir string `json:"db_dir"`     // path to data directory

	// admin mode. allows for expanded permissions when working with
	// the internal sfs file systems.
	AdminMode bool   `json:"admin_mode"`
	Admin     string `json:"admin"`
	AdminKey  string `json:"admin_key"`

all information about user file metadata  are saved in the database.
the above fields are saved as a json file.
*/
func (s *Service) SaveState() error {
	file, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal service state: %v", err)
	}

	sfName := fmt.Sprintf("sfs-state-%s.json", time.Now().Format("01-02-2006"))

	sfPath := filepath.Join(filepath.Join(s.SvcRoot, "state"))
	s.StateFile = filepath.Join(sfPath, sfName)

	return os.WriteFile(s.StateFile, file, 0644)
}

// returns the service run time in seconds
func (s *Service) RunTime() float64 {
	return time.Since(s.InitTime).Seconds()
}

// ------- new user service set up --------------------------------

// TODO: test!
//
// generate some base line meta data for this service instance.
// should generate a users.json file (which will keep track of active users),
// and a drives.json, containing info about each drive, its total size, its location,
// owner, init date, passwords, etc.
func (s *Service) GenBaseUserFiles(DrivePath string) {
	// create Drive directory
	if err := os.Mkdir(DrivePath, 0666); err != nil {
		log.Fatalf("[ERROR] failed to create Drive directory \n%v\n", err)
	}

	fileNames := []string{"user-info.json", "drive-info.json", "credentials.json"}
	for i := 0; i < len(fileNames); i++ {
		saveJSON(DrivePath, fileNames[i], make(map[string]interface{}))
	}
}

// Build a new privilaged Drive directory for a client on a Nimbus server
//
// Must be under /root/users/<username>
func (s *Service) AllocateDrive(name string, owner string) *files.Drive {
	usrs := filepath.Join(s.SvcRoot, "users")
	drivePath := filepath.Join(usrs, name)

	// generate service files
	s.GenBaseUserFiles(drivePath)

	// create new drive struct and save to DB
	// new drive id
	newRoot := files.NewRootDirectory(name, owner, drivePath)
	newDrive := files.NewDrive(files.NewUUID(), name, owner, drivePath, newRoot)

	return newDrive
}

// ------- user methods --------------------------------

// these will likely work with handlers

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
