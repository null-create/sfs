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

	// path for sfs service on the server
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

		// we don't set StateFile because we assume it
		// doesn't exist when NewService is called
		StateFile: "",
		UserDir:   filepath.Join(svcRoot, "users"),
		DbDir:     filepath.Join(svcRoot, "dbs"),

		// admin mode is optional.
		// these are standard default values
		AdminMode: false,
		Admin:     "admin",
		AdminKey:  "default",

		Users: make(map[string]*auth.User),
	}
}

// ------- init ---------------------------------------

func setAdmin(svc *Service) {
	cfg := ServerConfig()
	svc.AdminMode = true
	svc.Admin = cfg.Server.Admin
	svc.AdminKey = cfg.Server.AdminKey
}

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
		// NOTE: this might not be the most *current* version,
		// just the one that's present at the moment.
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
	// load state file and unmarshal into service struct
	file, err := os.ReadFile(sfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %v", err)
	}

	svc := &Service{}
	if err := json.Unmarshal(file, svc); err != nil {
		return nil, fmt.Errorf("failed unmarshal service state file: %v", err)
	}
	svc.StateFile = sfPath
	svc.InitTime = time.Now().UTC()

	return svc, nil
}

// populate svc.Users map from users database
func loadUsers(svc *Service) (*Service, error) {
	q := db.NewQuery(filepath.Join(svc.DbDir, "users"), false)
	usrs, err := q.GetUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user data from Users database: %v", err)
	}
	// NOTE: this assumes the physical files for each
	// user being loaded are already present. calling svc.AddUser()
	// will allocate a new drive service with new base files.
	for _, u := range usrs {
		if _, exists := svc.Users[u.ID]; !exists {
			svc.Users[u.ID] = u
		} else {
			log.Printf("[DEBUG] user %v already exists", u.ID)
		}
	}
	return svc, nil
}

// initialize a new sfs service from either a state file/dbs or
// create a new service from scratch.
func Init(new bool, admin bool) (*Service, error) {
	c := ServiceConfig()
	if !new {
		// ---- load from state file and dbs
		svc, err := SvcLoad(c.ServiceRoot, false)
		if err != nil {
			return nil, fmt.Errorf("[ERROR] failed to load service config: %v", err)
		}
		if admin {
			setAdmin(svc)
		}
		return svc, nil
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
func SvcInit(svcRoot string, debug bool) (*Service, error) {
	// ------- make root service directory (wherever it should located)
	log.Print("creating root service directory...")
	if err := os.Mkdir(svcRoot, 0644); err != nil {
		return nil, fmt.Errorf("[ERROR] failed to make service root directory: %v", err)
	}

	//-------- create top-level service directories
	log.Print("creating service subdirectories...")
	svcPaths := []string{
		filepath.Join(svcRoot, "users"),
		filepath.Join(svcRoot, "state"),
		filepath.Join(svcRoot, "dbs"),
	}
	for _, p := range svcPaths {
		if err := os.Mkdir(p, 0644); err != nil {
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
	svc := NewService(svcRoot)
	if err := svc.SaveState(); err != nil {
		return nil, fmt.Errorf("[ERROR] %v", err)
	}
	log.Print("all set :)")
	return svc, nil
}

//   - are the databases present?
//   - is the statefile present?
//
// if not, raise an error
func preChecks(svcRoot string) error {
	entries, err := os.ReadDir(svcRoot)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return fmt.Errorf("no statefile, user, or database directories found! \n%v", err)
	}
	for _, entry := range entries {
		if entry.Name() == "state" || entry.Name() == "dbs" {
			if isEmpty(filepath.Join(svcRoot, entry.Name())) {
				return fmt.Errorf("%s directory is empty \n%v", entry.Name(), err)
			}
		}
	}
	return nil
}

// intialize sfs service from a state file
func svcLoad(sfPath string) (*Service, error) {
	svc, err := loadStateFile(sfPath)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

// read in an external service state file
//
// populates users map through querying the users database
func SvcLoad(svcPath string, debug bool) (*Service, error) {
	// ensure (at least) the necessary dbs
	// and state files are present
	if err := preChecks(svcPath); err != nil {
		return nil, err
	}

	sfPath, err := findStateFile(svcPath)
	if err != nil {
		return nil, err
	}
	if sfPath == "" {
		return nil, fmt.Errorf("no state file found")
	}

	svc, err := svcLoad(sfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize service: %v", err)
	}
	// attempt to populate from users database
	// if state file had no user data
	if len(svc.Users) == 0 {
		log.Printf("[WARNING] state file had no user data. attempting to populate from users database...")
		_, err := loadUsers(svc)
		if err != nil {
			log.Fatalf("[ERROR] failed to retrieve user data: %v", err)
		}
	}

	// make sure we have a path to the db dir
	// and current state file for this session
	svc.DbDir = filepath.Join(svcPath, "dbs")
	svc.StateFile = sfPath
	return svc, nil
}

// ------ utils --------------------------------

// generate some base line meta data for this service instance.
// should generate a users.json file (which will keep track of active users),
// and a drives.json, containing info about each drive, its total size, its location,
// owner, init date, passwords, etc.
func GenBaseUserFiles(DrivePath string) {
	fileNames := []string{"user-info.json", "drive-info.json", "credentials.json"}
	for i := 0; i < len(fileNames); i++ {
		saveJSON(DrivePath, fileNames[i], make(map[string]interface{}))
	}
}

// build a new privilaged drive directory for a client on the sfs server
// with base state file info for user, drive, and user credentials json files
//
// must be under ../svcroot/users/<username>
func AllocateDrive(name string, owner string, svcRoot string) (*files.Drive, error) {
	// new user service file paths
	usrsDir := filepath.Join(svcRoot, "users")
	svcDir := filepath.Join(usrsDir, name)
	usrRoot := filepath.Join(svcDir, "root")

	// make a new user directory under svcRoot/users
	if err := os.Mkdir(svcDir, 0644); err != nil {
		return nil, err
	}

	// make an empty root directory (svcRoot/users/root)
	// for the user to store files
	if err := os.Mkdir(usrRoot, 0644); err != nil {
		return nil, err
	}

	rt := files.NewRootDirectory(name, owner, usrRoot)
	drv := files.NewDrive(files.NewUUID(), name, owner, svcDir, rt)

	// gen base files for this user
	GenBaseUserFiles(drv.DriveRoot)

	return drv, nil
}

// --------- service methods --------------------------------

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

	// TODO: find a better way to add hour:min:sec to state file name.
	sfName := fmt.Sprintf("sfs-state-%s.json", time.Now().Format("01-02-2006"))
	sfPath := filepath.Join(s.SvcRoot, "state")
	s.StateFile = filepath.Join(sfPath, sfName)

	return os.WriteFile(s.StateFile, file, 0644)
}

// NOTE: these will likely work with handlers

func (s *Service) TotalUsers() int {
	return len(s.Users)
}

func (s *Service) GetUser(id string) (*auth.User, error) {
	if usr, ok := s.Users[id]; ok {
		return usr, nil
	} else {
		return nil, fmt.Errorf("user %s not found", id)
	}
}

func (s *Service) GetUsers() map[string]*auth.User {
	if len(s.Users) == 0 {
		log.Printf("[DEBUG] no users found")
		return nil
	}
	return s.Users
}

// save to service instance and db
func (s *Service) addUser(u *auth.User, d *files.Drive) error {
	q := db.NewQuery(s.DbDir, false)
	if err := q.AddUser(u); err != nil {
		return fmt.Errorf("failed to add user to database: %v", err)
	}
	u.Drive = d
	s.Users[u.ID] = u
	return nil
}

// allocate a new service drive for a new user
func (s *Service) AddUser(u *auth.User) error {
	if _, exists := s.Users[u.ID]; !exists {
		// allocate new drive and base service files
		d, err := AllocateDrive(u.Name, u.Name, s.SvcRoot)
		if err != nil {
			return fmt.Errorf("failed to allocate new drive for user %s \n%v", u.ID, err)
		}
		// save to service instance and db
		if err := s.addUser(u, d); err != nil {
			return err
		}
	} else {
		log.Printf("[DEBUG] user (id=%s) already present", u.ID)
	}
	return nil
}

// remove a user and all their files and directories
func (s *Service) RemoveUser(id string) error {
	if usr, ok := s.Users[id]; ok {
		if len(usr.Drive.Root.Dirs) != 0 {
			if err := usr.Drive.Root.Clean(usr.Drive.Root.RootPath); err != nil {
				return fmt.Errorf(" unable to remove user and drive contents: %v", err)
			}
		}
		// remove from User directory map
		delete(s.Users, usr.Drive.ID)
	} else {
		return fmt.Errorf("user (id=%s) not found", id)
	}
	return nil
}
