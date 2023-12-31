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

	// path to users directory
	UserDir string `json:"user_dir"`

	// path to database directory
	DbDir string `json:"db_dir"`

	// db singleton connection
	Db *db.Query

	// admin mode. allows for expanded permissions when working with
	// the internal sfs file systems.
	AdminMode bool   `json:"admin_mode"`
	Admin     string `json:"admin"`
	AdminKey  string `json:"admin_key"`

	// key: user id, val is user struct.
	//
	// user structs contain a path string to the users Drive directory,
	// so this can be used for measuring disc size and executing
	// health checks
	Users map[string]*auth.User `json:"users"`

	// map of populated drives.
	// key == userID, val == *svc.Drive
	Drives map[string]*svc.Drive `json:"drives"`
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
		Db:        db.NewQuery(filepath.Join(svcRoot, "dbs"), true),

		// admin mode is optional.
		// these are standard default values
		AdminMode: false,
		Admin:     "admin",
		AdminKey:  "default",

		// initialize user and drives maps
		Users:  make(map[string]*auth.User),
		Drives: make(map[string]*svc.Drive),
	}
}

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
	sfDir := filepath.Join(s.SvcRoot, "state")

	// make sure we only have one state file at a time
	if err := s.cleanSfDir(sfDir); err != nil {
		return err
	}

	// marshal state instance and write out
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal service state: %v", err)
	}
	sfName := fmt.Sprintf("sfs-state-%s.json", time.Now().UTC().Format("2006-01-02T15-04-05"))
	s.StateFile = filepath.Join(sfDir, sfName)

	return os.WriteFile(s.StateFile, data, svc.PERMS)
}

// remove previous state file(s) before writing out.
// we only want the most recent one available at a time.
func (s *Service) cleanSfDir(sfDir string) error {
	if entries, err := os.ReadDir(sfDir); err == nil {
		for _, entry := range entries {
			sf := filepath.Join(sfDir, entry.Name())
			if err := os.Remove(sf); err != nil {
				return err
			}
		}
	} else {
		log.Printf("[WARNING] failed to remove previous state file(s): %v", err)
	}
	return nil
}

// ------- init ---------------------------------------

// create a new SFS service with databases.
// requires a configured .env file prior to running.
func Build() error {
	c := ServiceConfig()
	_, err := SvcInit(c.SvcRoot)
	if err != nil {
		return fmt.Errorf("failed to initialize new SFS service: %v", err)
	}
	return nil
}

// initialize a new server-side sfs service from either a state file/dbs or
// create a new service from scratch.
func Init(new bool, admin bool) (*Service, error) {
	c := ServiceConfig()
	if !new {
		// load from state file and dbs
		svc, err := SvcLoad(c.SvcRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to load service config: %v", err)
		}
		if admin {
			setAdmin(svc)
		}
		return svc, nil
	} else {
		// initialize new sfs service
		svc, err := SvcInit(c.SvcRoot)
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
	cfg := ServerConfig()
	svc.AdminMode = true
	svc.Admin = cfg.Admin
	svc.AdminKey = cfg.AdminKey
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
		log.Print("[WARNING] no users found in Users database")
		return svc, nil
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

// loads each drive from the database.
func loadDrives(svc *Service) (*Service, error) {
	drives, err := svc.Db.GetDrives()
	if err != nil {
		return svc, fmt.Errorf("failed to query database for drives: %v", err)
	}
	if len(drives) == 0 {
		log.Print("[INFO] no drives found")
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
	// successive service load quicker.
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
	log.Print("creating root service directory...")
	if err := os.Mkdir(svcRoot, 0644); err != nil {
		return nil, fmt.Errorf("failed to make service root directory: %v", err)
	}

	//create top-level service directories
	log.Print("creating service subdirectories...")
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
	log.Print("creating service databases...")
	if err := db.InitDBs(svcPaths[2]); err != nil {
		return nil, fmt.Errorf("failed to initialize service databases: %v", err)
	}

	// create new service instance and save initial state
	log.Print("initializng new service instance...")
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

	log.Print("all set :)")
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
		log.Printf("[WARNING] state file had no user data. attempting to populate from users database...")
		_, err := loadUsers(svc)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve user data: %v", err)
		}
	}
	if len(svc.Drives) == 0 {
		log.Printf("[WARNING] state file had no drive data. attempting to populate from drives database...")
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

// --------- drives --------------------------------

// check for whether a drive exists. does not check database.
func (s *Service) DriveExists(driveID string) bool {
	if _, exists := s.Drives[driveID]; exists {
		return true
	}
	return false
}

// Discover populates the given root directory with the users file and
// sub directories, updates the database as it does so, and returns
// the the directory object when finished.
//
// This should ideally be used for starting a new sfs service in a
// users root directly that already has files and/or subdirectories.
func (s *Service) Discover(root *svc.Directory) (*svc.Directory, error) {
	// traverse users SFS file system and populate internal structures
	root = root.Walk()
	// send everything to the database
	files := root.WalkFs()
	for _, file := range files {
		if err := s.Db.AddFile(file); err != nil {
			return nil, fmt.Errorf("failed to add file to database: %v", err)
		}
	}
	dirs := root.WalkDs()
	for _, d := range dirs {
		if err := s.Db.AddDir(d); err != nil {
			return nil, fmt.Errorf("failed to add directory to database: %v", err)
		}
	}
	// add root directory itself
	if err := s.Db.AddDir(root); err != nil {
		return nil, fmt.Errorf("failed to add root directory to database: %v", err)
	}
	return root, nil
}

// Populate() populates a drive's root directory with all the users
// files and subdirectories by searching the DB with the name
// of each file or directory Populate() discoveres as it traverses the
// users SFS filesystem.
//
// Note that Populate() ignores files and subdirectories it doesn't find in the
// database as its traversing the file system. This may or may not be a good thing.
func (s *Service) Populate(root *svc.Directory) *svc.Directory {
	if root.Path == "" {
		log.Print("[WARNING] can't traverse directory without a path")
		return nil
	}
	if root.IsNil() {
		log.Printf(
			"[WARNING] can't traverse directory with empty or nil maps: \nfiles=%v dirs=%v",
			root.Files, root.Dirs,
		)
		return nil
	}
	return s.populate(root)
}

func (s *Service) populate(dir *svc.Directory) *svc.Directory {
	entries, err := os.ReadDir(dir.Path)
	if err != nil {
		log.Printf("[ERROR]: %v", err)
		return dir
	}
	if len(entries) == 0 {
		log.Printf("[INFO] dir (id=%s) has no entries: ", dir.ID)
		return dir
	}
	for _, entry := range entries {
		entryPath := filepath.Join(dir.Path, entry.Name())
		item, err := os.Stat(entryPath)
		if err != nil {
			log.Printf("[ERROR] could not get stat for entry %s \nerr: %v", entryPath, err)
			return dir
		}
		// add directory then recurse
		if item.IsDir() {
			subDir, err := s.Db.GetDirectoryByName(item.Name())
			if err != nil {
				log.Printf("[ERROR] could not get directory from db: %v \nerr: %v", item.Name(), err)
				continue
			}
			if subDir == nil {
				continue // not found
			}
			subDir = s.populate(subDir)
			dir.AddSubDir(subDir)
		} else { // add file
			file, err := s.Db.GetFileByName(item.Name())
			if err != nil {
				log.Printf("[ERROR] could not get file (%s) from db: %v", item.Name(), err)
				continue
			}
			if file == nil {
				continue // not found
			}
			dir.AddFile(file)
		}
	}
	return dir
}

// recursively descends the drive's directory tree and compares what it
// finds to what is in the database, adding new items as it goes. generates
// a new root directory object and attaches it to the drive.
func (s *Service) RefreshDrive(driveID string) error {
	if s.DriveExists(driveID) {
		// get current full drive state
		drive, err := s.LoadDrive(driveID)
		if err != nil {
			return fmt.Errorf("failed to load drive: %v", err)
		}
		// refresh root against the database and create a new root object
		newRoot := s.refreshDrive(drive.Root)
		// clear old contents from memory then add new root
		drive.Root.Clear(drive.Root.Key)
		drive.Root = newRoot
		s.Drives[drive.ID] = drive
		if err := s.SaveState(); err != nil {
			return fmt.Errorf("failed to save state file: %v", err)
		}
	} else {
		return fmt.Errorf("drive (id=%s) not found", driveID)
	}
	return nil
}

func (s *Service) refreshDrive(dir *svc.Directory) *svc.Directory {
	entries, err := os.ReadDir(dir.Path)
	if err != nil {
		log.Printf("[ERROR]: %v", err)
		return dir
	}
	if len(entries) == 0 {
		log.Printf("[INFO] dir (id=%s) has no entries: ", dir.ID)
		return dir
	}
	for _, entry := range entries {
		entryPath := filepath.Join(dir.Path, entry.Name())
		item, err := os.Stat(entryPath)
		if err != nil {
			log.Printf("[ERROR] could not get stat for entry %s \nerr: %v", entryPath, err)
			return dir
		}
		// add directory then recurse
		if item.IsDir() {
			subDir, err := s.Db.GetDirectoryByName(item.Name())
			if err != nil {
				log.Printf("[ERROR] could not get directory from db: %v \nerr: %v", item.Name(), err)
				continue
			}
			// new directory
			if subDir == nil {
				subDir = svc.NewDirectory(item.Name(), dir.OwnerID, dir.DriveID, entryPath)
				if err := s.Db.AddDir(subDir); err != nil {
					log.Printf("[ERROR] could not add directory (%s) to db: %v", item.Name(), err)
					continue
				}
				subDir = s.refreshDrive(subDir)
				dir.AddSubDir(subDir)
			}
		} else {
			file, err := s.Db.GetFileByName(item.Name())
			if err != nil {
				log.Printf("[ERROR] could not get file (%s) from db: %v", item.Name(), err)
				continue
			}
			// new file
			if file == nil {
				newFile := svc.NewFile(item.Name(), dir.DriveID, dir.OwnerID, filepath.Join(item.Name(), dir.Path))
				if err := s.Db.AddFile(newFile); err != nil {
					log.Printf("[ERROR] could not add file (%s) to db: %v", item.Name(), err)
					continue // TEMP until there's a better way to handle this error
				}
				dir.AddFile(newFile)
			}
		}
	}
	return dir
}

// attempts to retrieve a drive from the drive map. if not found,
// will check database and if a drive is found, will add it to the map,
// update the service state file, and return the drive with the root
// directory populated. if no drive (or drive root) is found,
// then will return nil.
func (s *Service) GetDrive(driveID string) *svc.Drive {
	if drive, exists := s.Drives[driveID]; exists {
		return drive
	} else {
		log.Printf("[WARNING] drive %s not found in drive map. attempting to load from database", driveID)
		drive, err := s.Db.GetDrive(driveID)
		if err != nil {
			return nil
		}
		if drive == nil {
			return nil
		}
		root, err := s.Db.GetDirectory(drive.RootID)
		if err != nil {
			return nil
		}
		drive.Root = s.Populate(root)
		s.Drives[drive.ID] = drive
		if err := s.SaveState(); err != nil {
			log.Printf("[WARNING] failed to save state while updating drive map: %v", err)
		}
		return drive
	}
}

// save drive state to DB
func (s *Service) SaveDrive(drv *svc.Drive) error {
	if err := s.Db.UpdateDrive(drv); err != nil {
		return fmt.Errorf("failed to update drive in database: %v", err)
	}
	if err := drv.SaveState(); err != nil {
		return fmt.Errorf("failed to save drive state: %v", err)
	}
	return nil
}

// Populates drive's root directory with and all the users
// subdirectories and files from the database into memory.
// Should be followed by a call to Drive.Root.Clear() (not clean!)
// to clear the drive's internal data structures after loading.
func (s *Service) LoadDrive(driveID string) (*svc.Drive, error) {
	if drive, exists := s.Drives[driveID]; exists {
		drive.Root = s.Populate(drive.Root)
		return drive, nil
	} else {
		return nil, fmt.Errorf("drive %s not found", driveID)
	}
}

// remove a drive and all its files and directories, as well
// as its info from the database
func (s *Service) RemoveDrive(driveID string) error {
	drv := s.GetDrive(driveID)
	if drv == nil {
		log.Printf("[INFO] drive %s not found", driveID)
		return nil
	}
	// remove drive physical files/directories
	if err := Clean(drv.Root.Path); err != nil {
		return fmt.Errorf("failed to remove drives files and directories: %v", err)
	}
	// remove all files and directories from the database
	files := drv.Root.GetFiles()
	for _, f := range files {
		if err := s.Db.RemoveFile(f.ID); err != nil {
			return err
		}
	}
	dirs := drv.Root.GetSubDirs()
	for _, d := range dirs {
		if err := s.Db.RemoveDirectory(d.ID); err != nil {
			return err
		}
	}
	// remove root itself
	if err := s.Db.RemoveDirectory(drv.Root.ID); err != nil {
		return err
	}
	// remove drive from db
	if err := s.Db.RemoveDrive(driveID); err != nil {
		return err
	}
	// remove from drives map and save state
	delete(s.Drives, driveID)
	if err := s.SaveState(); err != nil {
		return err
	}
	log.Printf("[INFO] drive %s removed", driveID)
	return nil
}

// add drive and all files and subdirectoris to the service instance
func (s *Service) AddDrive(drv *svc.Drive) error {
	if drv.Root == nil {
		return fmt.Errorf("drive does not have root directory")
	}
	// add all files and subdirectories to the database
	files := drv.Root.WalkFs()
	for _, f := range files {
		if err := s.Db.AddFile(f); err != nil {
			return err
		}
	}
	subDirs := drv.Root.WalkDs()
	for _, d := range subDirs {
		if err := s.Db.AddDir(d); err != nil {
			return err
		}
	}
	// add drive and root dir to db
	if err := s.Db.AddDrive(drv); err != nil {
		return err
	}
	if err := s.Db.AddDir(drv.Root); err != nil {
		return err
	}
	s.Drives[drv.ID] = drv
	if err := s.SaveState(); err != nil {
		log.Printf("[WARNING] failed to save state file: %v", err)
	}
	return nil
}

// take a given drive instance and update db. does not traverse
// file systen for any other changes, only deals with drive metadata.
// use service.RefreshDrive(driveID) to do a complete refresh of a given drive and its
// file system.
func (s *Service) UpdateDrive(drv *svc.Drive) error {
	if s.DriveExists(drv.ID) {
		if err := s.Db.UpdateDrive(drv); err != nil {
			return err
		}
		s.Drives[drv.ID] = drv
		if err := s.SaveState(); err != nil {
			return fmt.Errorf("failed to save state file: %v", err)
		}
	} else {
		return fmt.Errorf("drive (id=%s) not found", drv.ID)
	}
	return nil
}

// --------- users --------------------------------

func (s *Service) TotalUsers() int {
	return len(s.Users)
}

// checks service instance and user db for whether a user exists
func (s *Service) UserExists(userID string) bool {
	if _, exists := s.Users[userID]; exists {
		// make sure they exist in the DB too
		exists, err := s.Db.UserExists(userID)
		if err != nil {
			log.Fatalf("failed to check user database: %v", err)
		}
		return exists
	} else {
		// last shot, try the DB
		exists, err := s.Db.UserExists(userID)
		if err != nil {
			log.Fatalf("failed to check user database: %v", err)
		}
		return exists
	}
}

// generate new user instance, and create drive and other base files
func (s *Service) addUser(user *auth.User) error {
	// check to see if this user already has a drive
	if dID, err := s.Db.GetDriveID(user.ID); err != nil {
		return fmt.Errorf("failed to get drive ID: %v", err)
	} else if dID != "" {
		return fmt.Errorf("user (%s) already has a drive (%s): ", user.ID, dID)
	}

	// allocate new drive and base service files
	newDrive, err := svc.AllocateDrive(user.Name, user.Name, s.SvcRoot)
	if err != nil {
		return fmt.Errorf("failed to allocate new drive for %s (id=%s) \n%v", user.Name, user.ID, err)
	}
	user.DriveID = newDrive.ID
	user.DrvRoot = newDrive.Root.Path

	// save info to the db, then update service user map
	if err := s.Db.AddUser(user); err != nil {
		return fmt.Errorf("failed to add user to database: %v", err)
	}
	if err := s.Db.AddDrive(newDrive); err != nil {
		return fmt.Errorf("failed to add drive to database: %v", err)
	}
	if err := s.Db.AddDir(newDrive.Root); err != nil {
		return fmt.Errorf("failed to add drive root directory to database: %v", err)
	}
	s.Drives[newDrive.ID] = newDrive
	s.Users[user.ID] = user
	if err := s.SaveState(); err != nil {
		log.Printf("[WARNING] failed to save state file: %v", err)
	}
	return nil
}

// allocate a new service drive for a new user.
//
// creates a new service drive, adds the user to the
// user database and service instance, and adds the
// drive to the drive database.
func (s *Service) AddUser(newUser *auth.User) error {
	if _, exists := s.Users[newUser.ID]; !exists {
		if err := s.addUser(newUser); err != nil {
			return err
		}
		log.Printf("[INFO] added user (name=%s id=%s)", newUser.Name, newUser.ID)
		if err := s.SaveState(); err != nil {
			log.Printf("[WARNING] failed to save service state: %s", err)
			return nil
		}
	} else {
		return fmt.Errorf("user (id=%s) already exists", newUser.ID)
	}
	return nil
}

// remove users's drive and all files and directories within,
// as well as all drive information from the database
func (s *Service) removeUser(driveID string) error {
	if d, err := s.Db.GetDrive(driveID); err == nil {
		if d != nil && d.DriveRoot != "" {
			// get root directory for this drive
			root, err := s.Db.GetDirectory(d.RootID)
			if err != nil {
				return err
			}
			if d.Root == nil {
				d.Root = root
			}
			// remove all files and directories
			if err := d.Root.Clean(d.DriveRoot); err != nil {
				return err
			}
			//remove users root dir itself
			if err := os.RemoveAll(d.DriveRoot); err != nil {
				return err
			}
			// remove drive from database
			if err := s.Db.RemoveDrive(driveID); err != nil {
				return err
			}
			// remove root directory from db
			if err := s.Db.RemoveDirectory(root.ID); err != nil {
				return err
			}
			log.Printf("[INFO] drive (id=%s) removed", driveID)
		} else {
			log.Printf("[DEBUG] drive (id=%s) not found", driveID)
		}
	} else {
		return err
	}
	return nil
}

// remove a user and all their files and directories
func (s *Service) RemoveUser(userID string) error {
	if usr, exists := s.Users[userID]; exists {
		// remove users drive and all files/directories
		if err := s.removeUser(usr.DriveID); err != nil {
			return err
		}
		// remove user from database
		if err := s.Db.RemoveUser(usr.ID); err != nil {
			return err
		}
		// delete from service instance
		delete(s.Users, usr.ID)
		log.Printf("[INFO] user (id=%s) removed", userID)
		if err := s.SaveState(); err != nil {
			return err
		}
	} else {
		log.Printf("[WARNING] user (id=%s) not found", userID)
	}
	return nil
}

// find a user. if not in the instance, then it will query the database.
//
// returns nil if user isn't found
func (s *Service) FindUser(userId string) (*auth.User, error) {
	if u, exists := s.Users[userId]; !exists {
		u, err := s.Db.GetUser(userId)
		if err != nil {
			return nil, err
		}
		if u == nil {
			log.Printf("[INFO] user (id=%s) not found", userId)
			return nil, nil
		}
		s.Users[u.ID] = u // add to the map since we didn't find it initially
		return u, nil
	} else {
		return u, nil
	}
}

func (s *Service) updateUser(user *auth.User) error {
	if err := s.Db.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to update user in database: %v", err)
	}
	s.Users[user.ID] = user
	log.Printf("[INFO] user (id=%s) updated", user.ID)
	return nil
}

// update user info
func (s *Service) UpdateUser(user *auth.User) error {
	if _, exists := s.Users[user.ID]; exists {
		u, err := s.Db.GetUser(user.ID)
		if err != nil {
			return err
		}
		if u != nil { // user exists, update it
			if err := s.updateUser(user); err != nil {
				return err
			}
			log.Printf("[INFO] user (id=%s) updated", user.ID)
			if err = s.SaveState(); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("user %s not found in user database", user.ID)
		}
	} else {
		return fmt.Errorf("user %s not found", user.ID)
	}
	return nil
}

// ---------- files --------------------------------

// find a file in the database
func (s *Service) FindFile(driveID string, fileID string) (*svc.File, error) {
	drive := s.GetDrive(driveID)
	if drive == nil {
		return nil, fmt.Errorf("drive id=%s not found", driveID)
	}
	file := drive.GetFile(fileID)
	if file == nil {
		return nil, fmt.Errorf("file %s not found", fileID)
	}
	return file, nil
}

// get all files for this user
func (s *Service) FindFiles(driveID string, fileIds []string) (map[string]*svc.File, error) {
	drive := s.GetDrive(driveID)
	if drive == nil {
		return nil, fmt.Errorf("drive id=%s not found", driveID)
	}
	files := drive.GetFiles()
	if len(files) == 0 {
		log.Printf("[INFO] no files in drive id=%s", driveID)
	}
	return files, nil
}

// add a file to the service.
func (s *Service) AddFile(dirID string, file *svc.File) error {
	drive := s.GetDrive(file.DriveID)
	if drive == nil {
		return fmt.Errorf("drive id=%s not found", file.DriveID)
	}
	if err := drive.AddFile(dirID, file); err != nil {
		return fmt.Errorf("failed to add file to drive: %v", err)
	}
	// add file to the database
	if err := s.Db.AddFile(file); err != nil {
		return fmt.Errorf("failed to add file to database: %v", err)
	}
	if err := s.SaveState(); err != nil {
		log.Printf("[WARNING] failed to save state: %v", err)
	}
	return nil
}

// update a file in the service
func (s *Service) UpdateFile(file *svc.File, data []byte) error {
	drive := s.GetDrive(file.DriveID)
	if drive == nil {
		return fmt.Errorf("drive id=%s not found", file.DriveID)
	}
	if err := drive.Root.UpdateFile(file, data); err != nil {
		return err
	}
	if err := s.Db.UpdateFile(file); err != nil {
		return err
	}
	if err := s.SaveState(); err != nil {
		log.Printf("[WARNING] failed to save state: %v", err)
	}
	return nil
}

// delete a file in the service. uses the users drive to delete the file.
// removes physical file and updates database.
func (s *Service) DeleteFile(file *svc.File) error {
	drive := s.GetDrive(file.DriveID)
	if drive == nil {
		return fmt.Errorf("drive (id=%s) not found", file.DriveID)
	}
	if err := drive.RemoveFile(file.DirID, file); err != nil {
		return fmt.Errorf("failed to remove %s (id=%s)s from drive: %v", file.Name, file.ID, err)
	}
	if err := s.Db.RemoveFile(file.ID); err != nil {
		return fmt.Errorf("failed to remove %s (id=%s) from database: %v", file.Name, file.ID, err)
	}
	if err := s.SaveState(); err != nil {
		log.Printf("[WARNING] failed to save state: %v", err)
	}
	return nil
}

// --------- directories --------------------------------

// find a directory in the database. does not populate with files or subdirectories,
// just returns metadata.
func (s *Service) FindDir(dirID string) (*svc.Directory, error) {
	dir, err := s.Db.GetDirectory(dirID)
	if err != nil {
		return nil, err
	}
	return dir, nil
}

// add a sub-directory to the given drive directory.
// makes a physical directory for this new directory object
// and updates the database.
func (s *Service) NewDir(driveID string, destDirID string, newDir *svc.Directory) error {
	drive := s.GetDrive(driveID)
	if drive == nil {
		return fmt.Errorf("drive %s not found", driveID)
	}
	if !drive.IsLoaded {
		drive.Root = s.Populate(drive.Root)
	}
	if err := drive.AddSubDir(destDirID, newDir); err != nil {
		return err
	}
	if err := s.Db.AddDir(newDir); err != nil {
		return err
	}
	return nil
}

// remove a physical directory from a user's drive service.
// use with caution!
//
// it's assumed dirID is a sub-directory within the drive, and not
// the drives root directory itself.
func (s *Service) RemoveDir(driveID string, dirID string) error {
	drive := s.GetDrive(driveID)
	if drive == nil {
		return fmt.Errorf("drive %s not found", driveID)
	}
	if !drive.IsLoaded {
		drive.Root = s.Populate(drive.Root)
	}
	if err := drive.RemoveDir(dirID); err != nil {
		return fmt.Errorf("failed to remove dir %s: %v", dirID, err)
	}
	if err := s.Db.RemoveDirectory(dirID); err != nil {
		return fmt.Errorf("failed to remove directory from database: %v", err)
	}
	if err := s.Db.UpdateDir(drive.Root); err != nil {
		return fmt.Errorf("failed to update root directory: %v", err)
	}
	return nil
}

// update a directory within a drive.
func (s *Service) UpdateDir(driveID string, dir *svc.Directory) error {
	drive := s.GetDrive(driveID)
	if drive == nil {
		return fmt.Errorf("drive %s not found", driveID)
	}
	if !drive.IsLoaded {
		drive.Root = s.Populate(drive.Root)
	}
	if err := drive.UpdateDir(dir.ID, dir); err != nil {
		return fmt.Errorf("failed to update dir %s (id=%s): %v", dir.Name, dir.ID, err)
	}
	if err := s.Db.UpdateDir(dir); err != nil {
		return fmt.Errorf("failed update dir %s (id=%s) in database: %v", dir.Name, dir.ID, err)
	}
	return nil
}

// retrieves all directories available for user using the current drive state.
// returns nil if no directories are available.
func (s *Service) GetAllDirs(driveID string) ([]*svc.Directory, error) {
	drive := s.GetDrive(driveID)
	if !drive.IsLoaded {
		drive.Root = s.Populate(drive.Root)
	}
	d := drive.Root.GetSubDirs()
	if len(d) == 0 {
		log.Printf("[WARNING] no directories found for user %v", drive.OwnerID)
		return nil, nil
	}
	dirs := make([]*svc.Directory, 0, len(d))
	for _, sd := range d {
		dirs = append(dirs, sd)
	}
	return dirs, nil
}

// --------- sync --------------------------------

// retrieve a sync index for a given drive. used by the client
// to compare against and initiate a sync operation, if necessary
// will be used as the first step in a sync operation on the client side.
// sync index will be nil if not found.
func (s *Service) GetSyncIdx(driveID string) (*svc.SyncIndex, error) {
	drive := s.GetDrive(driveID)
	if drive == nil {
		return nil, fmt.Errorf("drive %s not found", driveID)
	}
	// build sync index if necessary
	if len(drive.SyncIndex.LastSync) == 0 || drive.SyncIndex == nil {
		log.Printf("[WARNING] sync index for %s not found. creating...", driveID)
		ownerID, err := drive.GetOwnerID()
		if err != nil {
			return nil, fmt.Errorf("failed to get drive owner ID: %v", err)
		}
		drive.SyncIndex = drive.Root.WalkS(svc.NewSyncIndex(ownerID))
		if err := drive.SaveState(); err != nil {
			return nil, fmt.Errorf("failed to save drive state: %v", err)
		}
	}
	return drive.SyncIndex, nil
}
