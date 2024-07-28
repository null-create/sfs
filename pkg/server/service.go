package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/db"
	logs "github.com/sfs/pkg/logger"
	svc "github.com/sfs/pkg/service"
)

/*
Server-side SFS service instance.
*/
type Service struct {
	// service ID and init time
	ID       string    `json:"id"`
	InitTime time.Time `json:"init_time"`

	// path for sfs service on the server
	SvcRoot string `json:"service_root"`

	// Service configs
	svcCfgs *SvcCfg `json:"svc_cfgs"`

	// path to state file
	StateFile string `json:"state_file"`

	// path to users directory
	UserDir string `json:"user_dir"`

	// path to database directory
	DbDir string `json:"db_dir"`

	// db singleton connection
	Db *db.Query `json:"db"`

	// logger
	log *logs.Logger `json:"log"`

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
	var id = auth.NewUUID()
	return &Service{
		ID:       id,
		InitTime: time.Now().UTC(),
		svcCfgs:  svcCfg,
		SvcRoot:  svcRoot,

		// we don't set StateFile because we assume it
		// doesn't exist when NewService is called
		StateFile: "",
		UserDir:   filepath.Join(svcRoot, "users"),
		DbDir:     filepath.Join(svcRoot, "dbs"),
		Db:        db.NewQuery(filepath.Join(svcRoot, "dbs"), true),
		log:       logs.NewLogger("Service", id),

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
			if err2 := os.Remove(sf); err2 != nil {
				return err2
			}
		}
	} else {
		s.log.Error("failed to remove previous state file(s): " + err.Error())
	}
	return nil
}

// --------- drives --------------------------------

// check for whether a drive exists. does not check database.
func (s *Service) HasDrive(driveID string) bool {
	if _, exists := s.Drives[driveID]; exists {
		return true
	}
	return false
}

// similar to HasDrive, but checks the DB if the drive isn't found in the
// instance map before giving up. if the drive is found in the db and wasn't
// in the instance map previous, it gets added to the map and the service
// state gets updated.
func (s *Service) DriveExists(driveID string) bool {
	if _, exists := s.Drives[driveID]; !exists {
		drive, err := s.Db.GetDrive(driveID)
		if err != nil {
			s.log.Error("failed to get drive info from database: " + err.Error())
			return false
		}
		if drive != nil {
			// save this to the service instance since it wasn't there
			// before for some reason
			s.Drives[driveID] = drive
			if err := s.SaveState(); err != nil {
				s.log.Error("failed to update state file: " + err.Error())
			}
			return true
		}
		return false
	}
	return true
}

// Populate() populates a drive's root directory with all the users
// files and subdirectories by recursively traversing the users server-side file system
// and searching the DB with the name of each file or directory Populate() discoveres
//
// Note that Populate() ignores files and subdirectories it doesn't find in the
// database as its traversing the file system.
func (s *Service) Populate(root *svc.Directory) *svc.Directory {
	if root.Path == "" {
		s.log.Error("can't traverse directory without a path")
		return root
	}
	return s.populate(root)
}

func (s *Service) populate(dir *svc.Directory) *svc.Directory {
	entries, err := os.ReadDir(dir.ServerPath)
	if err != nil {
		s.log.Error("can't read directory: " + dir.ServerPath)
		return dir
	}
	if len(entries) == 0 {
		return dir
	}
	for _, entry := range entries {
		entryPath := filepath.Join(dir.ServerPath, entry.Name())
		item, err := os.Stat(entryPath)
		if err != nil {
			s.log.Error(fmt.Sprintf("could not get stat for entry %s: %v", entryPath, err))
			continue
		}
		// add directory then recurse
		if item.IsDir() {
			subDir, err := s.Db.GetDirectoryByPath(entryPath)
			if err != nil {
				s.log.Error(fmt.Sprintf("could not get directory (%s) from db: %s", item.Name(), err))
				continue
			}
			if subDir == nil {
				continue // not found
			}
			subDir = s.populate(subDir)
			dir.AddSubDir(subDir)
		} else { // add file
			file, err := s.Db.GetFileByPath(entryPath)
			if err != nil {
				s.log.Error(fmt.Sprintf("could not get file (%s) from db: %v", item.Name(), err))
				continue
			}
			if file == nil {
				continue // not found
			}
			if err := dir.AddFile(file); err != nil {
				s.log.Error(fmt.Sprintf("could not add file (%s) to db: %v", file.Name, err))
			}
		}
	}
	return dir
}

// attempts to retrieve a drive from the drive map.
// populates the drive if found.
func (s *Service) GetDrive(driveID string) *svc.Drive {
	if drive, exists := s.Drives[driveID]; exists {
		root, err := s.loadRoot(drive.RootID)
		if err != nil {
			s.log.Error(fmt.Sprintf("failed to load drive root %s: %v", drive.RootID, err))
		}
		drive.Root = root
		drive.IsLoaded = true
		return drive
	}
	return nil
}

// save drive state to DB
func (s *Service) SaveDrive(drv *svc.Drive) error {
	if err := s.Db.UpdateDrive(drv); err != nil {
		return fmt.Errorf("failed to update drive in database: %v", err)
	}
	return nil
}

// load and populate the root directory
func (s *Service) loadRoot(rootID string) (*svc.Directory, error) {
	root, err := s.Db.GetDirectoryByID(rootID)
	if err != nil {
		return nil, fmt.Errorf("failed to get root directory: %v", err)
	}
	if root == nil {
		return nil, fmt.Errorf("root directory (id=%s) not found", rootID)
	}
	popRoot := s.Populate(root)
	return popRoot, nil
}

// Loads drive and root directory from the database, populates
// the root, and generates a new sync index.
func (s *Service) LoadDrive(driveID string) (*svc.Drive, error) {
	drive, err := s.Db.GetDrive(driveID)
	if err != nil {
		return nil, err
	}
	if drive == nil {
		return nil, fmt.Errorf("drive %s not found", driveID)
	}
	root, err := s.Db.GetDirectoryByID(drive.RootID)
	if err != nil {
		return nil, err
	}
	if root == nil {
		return nil, fmt.Errorf("no root directory found for drive %s", driveID)
	}
	drive.Root = root

	// add all users directories
	dirs, err := s.Db.GetDirsByDriveID(driveID)
	if err != nil {
		return nil, fmt.Errorf("failed to load users directories: %v", err)
	}
	drive.Root.AddSubDirs(dirs)
	s.log.Log(logs.INFO, fmt.Sprintf("added %d directories to drive id=%s", len(dirs), driveID))

	// add all users files
	files, err := s.Db.GetFilesByDriveID(driveID)
	if err != nil {
		return nil, fmt.Errorf("failed to load users files: %v", err)
	}
	drive.Root.AddFiles(files)
	s.log.Log(logs.INFO, fmt.Sprintf("added %d files to drive id=%s", len(files), driveID))

	// populate the root directory and generate a new sync index
	drive.Root = s.Populate(root)
	drive.SyncIndex = svc.BuildRootSyncIndex(drive.Root)

	// save to service instance
	s.Drives[drive.ID] = drive
	if err := s.SaveState(); err != nil {
		return nil, fmt.Errorf("[ERROR] failed to save service state while updating drive map: %v", err)
	}
	return drive, nil
}

// add a new drive to the service instance.
// saves the drives root directory and the drive itself to the server's databases.
func (s *Service) AddDrive(drv *svc.Drive) error {
	// create a server-side user root directory for this drive.
	// check if this drive is already registered first.
	root, err := s.Db.GetDirectoryByID(drv.RootID)
	if err != nil {
		return err
	}
	if root != nil {
		return fmt.Errorf("drive for '%s' (drv-id=%s) is already registered", drv.OwnerName, drv.ID)
	}
	// create root from existing client-side drive
	// info so we can register this new drive
	root = svc.NewRootDirectory("root", drv.OwnerID, drv.ID, drv.RootPath)
	root.ServerPath = filepath.Join(s.SvcRoot, "users", drv.OwnerName)
	drv.Root = root
	drv.RootID = root.ID
	drv.RootPath = root.ServerPath

	// add drive and root dir to db
	if err := s.Db.AddDir(drv.Root); err != nil {
		return err
	}
	if err := s.Db.AddDrive(drv); err != nil {
		return err
	}

	// allocate new physical drive directories
	if err := svc.AllocateDrive(drv.OwnerName, s.SvcRoot); err != nil {
		s.log.Error(fmt.Sprintf(
			"failed to allocate new drive for %s (id=%s): %v",
			drv.OwnerName, drv.OwnerID, err),
		)
		return fmt.Errorf(
			"failed to allocate new drive for %s (id=%s): %v",
			drv.OwnerName, drv.OwnerID, err,
		)
	}

	// initalize sync index and save to service instance
	drv.SyncIndex = svc.NewSyncIndex(drv.OwnerID)
	s.Drives[drv.ID] = drv
	if err := s.SaveState(); err != nil {
		return fmt.Errorf("failed to save state: %v", err)
	}
	return nil
}

// take a given drive instance and update db. does not traverse
// file systen for any other changes, only deals with drive metadata.
func (s *Service) UpdateDrive(drv *svc.Drive) error {
	if s.HasDrive(drv.ID) {
		if err := s.Db.UpdateDrive(drv); err != nil {
			return err
		}
		s.Drives[drv.ID] = drv
		if err := s.SaveState(); err != nil {
			return fmt.Errorf("failed to save state: %v", err)
		}
	} else {
		return fmt.Errorf("drive (id=%s) not found", drv.ID)
	}
	return nil
}

// remove a drive and all the users files and directories, as well
// as its info from the database.
func (s *Service) RemoveDrive(driveID string) error {
	drv := s.GetDrive(driveID)
	if drv == nil {
		s.log.Info(fmt.Sprintf("drive %s not found", driveID))
		return nil
	}
	// remove drive physical files/directories
	if err := Clean(drv.Root.Path); err != nil {
		return fmt.Errorf("failed to remove drives files and directories: %v", err)
	}
	// remove all files and directories from the database
	files := drv.GetFilesMap()
	for _, f := range files {
		if err := s.Db.RemoveFile(f.ID); err != nil {
			return err
		}
	}
	dirs := drv.GetDirsMap()
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
	s.log.Info(fmt.Sprintf("drive (id=%s) removed", driveID))
	return nil
}

// --------- users --------------------------------

func (s *Service) TotalUsers() int { return len(s.Users) }

// checks service instance and user db for whether a user exists
func (s *Service) UserExists(userID string) bool {
	if _, ok := s.Users[userID]; ok {
		return true
	}
	return false
}

// generate new user instance, and create drive and other base files
func (s *Service) addUser(user *auth.User) error {
	// check to see if this user is already registered
	u, err := s.Db.GetUser(user.ID)
	if err != nil {
		return err
	}
	if u != nil {
		return fmt.Errorf("user '%s' (id=%s) is already registered", user.Name, user.ID)
	}
	if err := s.Db.AddUser(user); err != nil {
		return fmt.Errorf("failed to add %s (id=%s) to the user database: %v", user.Name, user.ID, err)
	}
	s.Users[user.ID] = user
	if err := s.SaveState(); err != nil {
		return err
	}
	return nil
}

// allocate a new service drive for a new user. used for first time set up.
//
// creates a new service drive, creates a new physical root
// directory on the server for the users files and directories,
// and adds the new drive, root, user, and all other necessary
// info to the database.
func (s *Service) AddUser(newUser *auth.User) error {
	if _, exists := s.Users[newUser.ID]; !exists {
		if err := s.addUser(newUser); err != nil {
			return err
		}
		s.log.Info(fmt.Sprintf("added user (name=%s id=%s)", newUser.Name, newUser.ID))
		return nil
	} else {
		return fmt.Errorf("user (id=%s) already exists", newUser.ID)
	}
}

// remove a user and all their files and directories
func (s *Service) RemoveUser(userID string) error {
	if usr, exists := s.Users[userID]; exists {
		// remove users drive and all files/directories
		if err := s.RemoveDrive(usr.DriveID); err != nil {
			return err
		}
		// remove user from database
		if err := s.Db.RemoveUser(usr.ID); err != nil {
			return err
		}
		// delete from service instance
		delete(s.Users, usr.ID)
		s.log.Info(fmt.Sprintf("user (id=%s) removed", userID))
		if err := s.SaveState(); err != nil {
			return err
		}
	} else {
		s.log.Warn(fmt.Sprintf("user (id=%s) not found", userID))
	}
	return nil
}

// find a user. if not in the instance, then it will query the database.
//
// returns nil if user isn't found
func (s *Service) GetUser(userId string) (*auth.User, error) {
	if u, exists := s.Users[userId]; !exists {
		// try the database before giving up
		u, err := s.Db.GetUser(userId)
		if err != nil {
			return nil, err
		}
		if u == nil {
			s.log.Info(fmt.Sprintf("user (id=%s) not found", userId))
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
	s.log.Info(fmt.Sprintf("user %s (id=%s) updated", user.Name, user.ID))
	return nil
}

// update user info
func (s *Service) UpdateUser(user *auth.User) error {
	if _, exists := s.Users[user.ID]; exists {
		if err := s.updateUser(user); err != nil {
			return err
		}
		if err := s.SaveState(); err != nil {
			return fmt.Errorf("failed to save service state: %v", err)
		}
	} else {
		// try DB before giving up
		u, err := s.Db.GetUser(user.ID)
		if err != nil {
			return err
		} else if u == nil {
			return fmt.Errorf("user (id=%s) not found", user.ID)
		}
		if err := s.updateUser(user); err != nil {
			return err
		}
		if err := s.SaveState(); err != nil {
			return fmt.Errorf("failed to save state: %v", err)
		}
	}
	return nil
}

// ---------- files --------------------------------

// find a file in the drive instance and return.
func (s *Service) GetFile(driveID string, fileID string) (*svc.File, error) {
	drive, err := s.LoadDrive(driveID)
	if err != nil {
		return nil, fmt.Errorf("failed to load drive: %v", err)
	}
	if drive == nil {
		return nil, fmt.Errorf("drive (id=%s) not found", driveID)
	}
	file := drive.GetFile(fileID)
	if file == nil {
		return nil, fmt.Errorf("file (id=%s) not found", fileID)
	}
	return file, nil
}

// get all file objects for this user.
func (s *Service) GetAllFiles(driveID string) (map[string]*svc.File, error) {
	drive, err := s.LoadDrive(driveID)
	if err != nil {
		return nil, fmt.Errorf("failed to load drive: %v", err)
	}
	if drive == nil {
		return nil, fmt.Errorf("drive (id=%s) not found", driveID)
	}
	files := drive.GetFilesMap()
	if len(files) == 0 {
		s.log.Info(fmt.Sprintf("no files in drive (id=%s)", driveID))
	}
	return files, nil
}

// generate a server-side path for a file or directory.
// path points the new item to the 'root' directory on the server
func (s *Service) buildServerRootPath(user string, itemName string) string {
	return filepath.Join(s.svcCfgs.SvcRoot, "users", user, "root", itemName)
}

// generate a server-side path for a file that has a server-side directory
func (s *Service) buildServerDirPath(user string, itemName string, dirPath string) string {
	return filepath.Join(s.svcCfgs.SvcRoot, "users", user, dirPath, itemName)
}

// add a new file to the service. creates the physical file,
// and updates internal service state.
func (s *Service) AddFile(dirID string, file *svc.File) error {
	drive, err := s.LoadDrive(file.DriveID)
	if err != nil {
		return fmt.Errorf("failed to load drive: %v", err)
	}
	if drive == nil {
		return fmt.Errorf("drive (id=%s) not found", file.DriveID)
	}
	// make sure the files parent directory exists on the server
	// first. if not, add to server-side sfs root.
	dir, err := s.Db.GetDirectoryByID(dirID)
	if err != nil {
		return err
	}
	// modify file.ServerPath to point to the server
	// side users root directory (or subdirectory if managed by the
	// server side root directory). whenever something gets
	// uploaded to the server we need to set a unique server path so we
	// can differentiate between client and server upload/download locations.
	// NOTE: client makes an additional call to retrieve this new path
	if dir == nil {
		// we're going to assign this file to root if the client side
		// parent directory isn't registered server-side yet.
		file.DirID = drive.Root.ID
		file.ServerPath = s.buildServerRootPath(drive.OwnerName, file.Name)
	} else {
		file.DirID = dir.ID
		file.ServerPath = s.buildServerDirPath(drive.OwnerName, file.Name, dir.ServerPath)
	}

	// create the (empty) physical file on the server side
	_, err = os.Create(file.ServerPath)
	if err != nil {
		return fmt.Errorf("failed to create empty file on server: %v", err)
	}

	// add any file contents to the server side
	if err := os.WriteFile(file.ServerPath, file.Content, svc.PERMS); err != nil {
		s.log.Error("failed to write file on server: " + err.Error())
	}

	// mark this as a server back up so we can access it
	// using the correct path on the server side
	file.MarkServerBackUp()

	// add file to drive service
	if err := drive.AddFile(file.DirID, file); err != nil {
		return fmt.Errorf("failed to add file to drive: %v", err)
	}
	// add file to the database
	if err := s.Db.AddFile(file); err != nil {
		return fmt.Errorf("failed to add file to database: %v", err)
	}
	if err := s.SaveState(); err != nil {
		return fmt.Errorf("failed to save state: %v", err)
	}
	return nil
}

// update a file in the service.
func (s *Service) UpdateFile(file *svc.File, data []byte) error {
	drive, err := s.LoadDrive(file.DriveID)
	if err != nil {
		return fmt.Errorf("failed to load drive: %v", err)
	}
	if drive == nil {
		return fmt.Errorf("drive (id=%s) not found", file.DriveID)
	}
	// make sure we update this file in the correct directory
	dir := drive.GetDir(file.DirID)
	if dir == nil {
		return fmt.Errorf("file's directory not found")
	}
	if err := dir.ModifyFile(file, data); err != nil {
		return err
	}
	if err := s.Db.UpdateFile(file); err != nil {
		return err
	}
	if err := s.SaveState(); err != nil {
		return fmt.Errorf("failed to save state: %v", err)
	}
	return nil
}

// deletes a file and updates the database.
func (s *Service) DeleteFile(file *svc.File) error {
	drive, err := s.LoadDrive(file.DriveID)
	if err != nil {
		return fmt.Errorf("failed to load drive: %v", err)
	}
	if drive == nil {
		return fmt.Errorf("drive (id=%s) not found", file.DriveID)
	}
	// remove file from the service.
	// NOTE: client side will have the original file moved to the client's recycle bin.
	if err := drive.RemoveFile(file.DirID, file); err != nil {
		return fmt.Errorf("failed to remove %s (id=%s)s from drive: %v", file.Name, file.ID, err)
	}
	if err := s.Db.RemoveFile(file.ID); err != nil {
		return fmt.Errorf("failed to remove %s (id=%s) from database: %v", file.Name, file.ID, err)
	}
	if err := s.SaveState(); err != nil {
		s.log.Error(fmt.Sprintf("failed to save state: %v", err))
	}
	// finally, remove the physical file on the server.
	// dont want users files to be left on the server after removing them
	// from the service.
	if err := os.Remove(file.ServerPath); err != nil {
		s.log.Error(fmt.Sprintf("failed to remove physical on the server: %v", err))
	}
	return nil
}

// --------- directories --------------------------------

// find a directory in the database. does not populate with files or subdirectories,
// just returns metadata.
func (s *Service) FindDir(driveID string, dirID string) (*svc.Directory, error) {
	drive := s.GetDrive(driveID)
	if drive == nil {
		return nil, fmt.Errorf("drive (id=%s) not found", driveID)
	}
	dir := drive.GetDir(dirID)
	if dir == nil {
		return nil, fmt.Errorf("directory (id=%s) not found", dirID)
	}
	return dir, nil
}

// add a sub-directory to the given drive directory
// and updates the database.
func (s *Service) NewDir(driveID string, destDirID string, newDir *svc.Directory) error {
	drive := s.GetDrive(driveID)
	if drive == nil {
		return fmt.Errorf("drive (id=%s) not found", driveID)
	}
	// check if the parent directory exists on the server. if not, add to root.
	var id string
	dir, err := s.Db.GetDirectoryByID(destDirID)
	if err != nil {
		return err
	}
	if dir != nil {
		id = dir.ID
		newDir.Parent = dir
	} else {
		id = drive.RootID
		newDir.Parent = drive.Root
	}
	// add directory to service.
	if err := drive.AddSubDir(id, newDir); err != nil {
		return err
	}
	if err := s.Db.AddDir(newDir); err != nil {
		return err
	}
	// finally, create the physical directory
	if err := os.MkdirAll(newDir.ServerPath, svc.PERMS); err != nil {
		return err
	}
	return nil
}

// remove a physical directory from a user's drive service.
// use with caution! will remove all children of this subdirectory
// as well.
//
// it's assumed dirID is a sub-directory within the drive, and not
// the drives root directory itself.
func (s *Service) RemoveDir(driveID string, dirID string) error {
	drive := s.GetDrive(driveID)
	if drive == nil {
		return fmt.Errorf("drive (id=%s) not found", driveID)
	}
	// just to be sure
	if dirID == drive.Root.ID {
		return fmt.Errorf("dir (id=%s) is drive root. cant remove root", dirID)
	}
	// get the directory to be removed, as well as all potential children.
	// we'll need to remove all of them from the DB.
	dir := drive.GetDir(dirID)
	if dir == nil {
		return fmt.Errorf("dir (id=%s) not found", dirID)
	}
	// remove all subdirs of this directory from the db
	subDirs := dir.GetDirMap()
	for _, subDir := range subDirs {
		if err := drive.RemoveDir(subDir.ID); err != nil {
			return err
		}
		if err := s.Db.RemoveDirectory(subDir.ID); err != nil {
			return err
		}
	}
	// remove all files from db
	files := dir.GetFiles()
	for _, file := range files {
		if err := drive.RemoveFile(file.DirID, file); err != nil {
			return err
		}
		if err := s.Db.RemoveFile(file.ID); err != nil {
			return err
		}
	}
	// remove directory itself from the service
	if err := s.Db.RemoveDirectory(dirID); err != nil {
		return fmt.Errorf("failed to remove directory from database: %v", err)
	}
	if err := s.Db.UpdateDir(drive.Root); err != nil {
		return fmt.Errorf("failed to update root directory: %v", err)
	}
	if err := drive.RemoveDir(dirID); err != nil {
		return fmt.Errorf("failed to remove dir %s: %v", dirID, err)
	}
	// lastly, remove the physical directory and all its subdirectories
	// don't want users files to remain on the server after they're done.
	if err := os.RemoveAll(dir.ServerPath); err != nil {
		return fmt.Errorf("failed to remove physical directory on server: %s", err)
	}
	return nil
}

// update a directory within a drive.
func (s *Service) UpdateDir(driveID string, dir *svc.Directory) error {
	drive := s.GetDrive(driveID)
	if drive == nil {
		return fmt.Errorf("drive (id=%s) not found", driveID)
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
	if drive == nil {
		return nil, fmt.Errorf("drive (id=%s) not found", driveID)
	}
	subDirs := drive.GetDirsMap()
	if len(subDirs) == 0 {
		s.log.Info(fmt.Sprintf("no directories found for user (id=%s)", drive.OwnerID))
		return nil, nil
	}
	dirs := make([]*svc.Directory, 0, len(subDirs))
	for _, sd := range subDirs {
		dirs = append(dirs, sd)
	}
	return dirs, nil
}

func (s *Service) MoveDir(driveID string, dirID string, destDirID string) error {
	drive := s.GetDrive(driveID)
	if drive == nil {
		return fmt.Errorf("drive (id=%s) not found", driveID)
	}

	// directory to move
	dir := drive.GetDir(dirID)
	if dir == nil {
		return fmt.Errorf("dir (id=%s) not found", dirID)
	}
	// parent of this directory
	// parent := dir.Parent
	// parent.RemoveSubDir(dir.ID)

	// TODO: directory.Copy(destPath)

	// directory to move to
	destDir := drive.GetDir(destDirID)
	if destDir == nil {
		return fmt.Errorf("dest dir (id=%s) not found", destDirID)
	}

	return nil
}

// --------- sync --------------------------------

// generate (or refresh) a drives sync index. returns nil if the
// drive, or the drive's root is not initialized or found.
func (s *Service) GenSyncIndex(driveID string) (*svc.SyncIndex, error) {
	drive := s.GetDrive(driveID)
	if drive == nil {
		return nil, fmt.Errorf("drive (id=%s) not found", driveID)
	}
	if drive.Root == nil {
		return nil, fmt.Errorf("drive (id=%s) root not found", drive.RootID)
	}
	drive.SyncIndex = svc.BuildRootSyncIndex(drive.Root)
	return drive.SyncIndex, nil
}

// retrieve a sync index for a given drive. used by the client
// to compare against and initiate a sync operation
// will be used as the first step in a sync operation on the client side.
// sync index will be nil if not found.
func (s *Service) GetSyncIdx(driveID string) (*svc.SyncIndex, error) {
	drive := s.GetDrive(driveID)
	if drive == nil {
		return nil, fmt.Errorf("drive (id=%s) not found", driveID)
	}
	if !drive.IsIndexed() {
		return nil, fmt.Errorf("drive (id=%s) is not indexed", driveID)
	}
	return drive.SyncIndex, nil
}

// refresh a drives updates map. used by clients during comparison operations
func (s *Service) RefreshUpdates(driveID string) (*svc.SyncIndex, error) {
	drive := s.GetDrive(driveID)
	if drive == nil {
		return nil, fmt.Errorf("drive (id=%s)s not found", driveID)
	}
	// drive should already be indexed
	if !drive.IsIndexed() {
		return nil, fmt.Errorf("drive (id=%s) has not been indexed", driveID)
	}
	drive.SyncIndex = svc.BuildRootToUpdate(drive.Root, drive.SyncIndex)
	return drive.SyncIndex, nil
}
