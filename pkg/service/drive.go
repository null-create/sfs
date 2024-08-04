package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sfs/pkg/logger"
)

// max size of a single drive (root directory) per user (10GB)
const MAX_SIZE int64 = 10e+9

/*
build a new privilaged drive directory for a client on the sfs server
with base state file info for user and drive json files

must be under ../svcroot/users/<username>

drives should have the following structure:

[client]
user/
|---root/      <---- user files & directories live here
|---state/
|   |---client-state-d-m-y-hh-mm-ss.json
|   recycled/     <---- "deleted" files & directories

[server]
users/
|----userA/
|    |----root/     <---- user files & directories
|    |----state/
|    |    |----drive-state-d-m-y-hh-mm-ss.json
|    |    recycled/     <---- "deleted" files & directories
|----userB/
(etc)

TODO:

	possibly store clients files on the server in a "flat" directory?
	doesn't have to match the client's local file system tree -- only be a
	repository of back ups. would make search time linear instead of whatever the walk()
	implementations in dirs.go are at currently.
*/
func AllocateDrive(name string, svcRoot string) error {
	// new user service file paths
	userRoot := filepath.Join(svcRoot, "users", name)
	// make each directory
	serviceDirs := []string{
		userRoot,
		filepath.Join(userRoot, "root"),
		filepath.Join(userRoot, "state"),
		filepath.Join(userRoot, "recycled"),
		filepath.Join(userRoot, "backups"),
	}
	for _, d := range serviceDirs {
		if err := os.Mkdir(d, PERMS); err != nil {
			return err
		}
	}
	return nil
}

/*
"Drives" are just abstractions of a protrected root directory,
managed by Simple File Sync, containing backups of a user's files and
other subdirectoriesto to facilitate synchronization across multiple devices.

Its basically just a directory containing some metadata about a users
drive size, state, and some security configurations, which itself
contains the user's "root" directory. Its this "root" directory where
all the users files live, in whatever arragement they end up using.

user/
|---root/    <---- the "drive." user's files & directories live here
|---state/
|   |---userID-d-m-y-hh-mm-ss.json

Drives may be realized as a filesystem on a user's current desktop,
laptop, dedicated hardrive within a desktop, or separate server.
*/
type Drive struct {
	ID        string `json:"drive_id"`
	OwnerName string `json:"owner_name"`
	OwnerID   string `json:"owner_id:"`

	// all three measured in Kb or Mb
	TotalSize int64 `json:"total_size"`
	UsedSpace int64 `json:"used_space"`
	FreeSpace int64 `json:"free_space"`

	// Security stuff
	Protected bool   `json:"protected"`
	Key       string `json:"-"`
	AuthType  string `json:"auth_type"`

	// Flag for whether Populate() has been called
	// with the drive's root directory. If so then the
	// drive's root directory will have its internal data structures
	// loaded and will make other calls to the users file contents possible.
	IsLoaded bool `json:"is_loaded"`

	// User's root directory & sync index
	RootPath   string         `json:"drive_root"` // location of the drive on physical server filesystem
	RootID     string         `json:"root_id"`
	Root       *Directory     `json:"-"` // ignored to avoid json cycle errors
	SyncIndex  *SyncIndex     `json:"sync_index"`
	Registered bool           `json:"registered"` // flag for whether this drive is registered with the SFS server
	Log        *logger.Logger `json:"-"`          // logging

	// folder for placing "deleted" files and directories
	RecycleBin string `json:"recycle_bin"`
}

var initLog = logger.NewLogger("DRIVE_INIT", "None")

func check(driveID string, ownerName string, ownerID string, rootPath string, root *Directory) bool {
	if driveID == "" || ownerName == "" || ownerID == "" || rootPath == "" || root == nil {
		initLog.Error("invalid drive parameters. none can be empty!")
		return false
	}
	return true
}

// creates a new drive service for a user. does not create new physical files
func NewDrive(driveID string, ownerName string, ownerID string, rootPath string, rootID string, root *Directory) *Drive {
	if !check(driveID, ownerName, ownerID, rootPath, root) {
		return nil
	}
	return &Drive{
		ID:         driveID,
		OwnerName:  ownerName,
		OwnerID:    ownerID,
		TotalSize:  MAX_SIZE,
		UsedSpace:  0,
		FreeSpace:  MAX_SIZE,
		Protected:  false,
		Key:        "default",
		RootPath:   rootPath,
		RootID:     rootID,
		Root:       root,
		RecycleBin: filepath.Join(root.Path, "recycle"),
		Log:        logger.NewLogger("DRIVE", driveID),
	}
}

// check whether this drive has an instantiated root directory.
func (d *Drive) HasRoot() bool { return d.Root != nil }

// check whether this drive is registered with the server.
func (d *Drive) IsRegistered() bool { return d.Registered }

// check whether the root directory has files and subdirectories
func (d *Drive) EmptyRoot() bool {
	return len(d.Root.Files) == 0 && len(d.Root.Dirs) == 0
}

// check whether the sync index has been initialized and populated.
func (d *Drive) IsIndexed() bool {
	return d.SyncIndex != nil && len(d.SyncIndex.LastSync) != 0
}

// save drive state to JSON format
func (d *Drive) ToJSON() ([]byte, error) {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return nil, err
	}
	return data, nil
}

func UnmarshalDriveString(data string) (*Drive, error) {
	drv := new(Drive)
	if err := json.Unmarshal([]byte(data), &drv); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dir data: %v", err)
	}
	return drv, nil
}

func (d *Drive) UpdateDriveSize(size int64) {
	d.UsedSpace += size
	d.FreeSpace -= size
}

// ------- security --------------------------------

func (d *Drive) Lock(password string) {
	if password != d.Key {
		d.Log.Info("wrong password")
	} else {
		d.Protected = true
	}
}

func (d *Drive) Unlock(password string) {
	if password != d.Key {
		d.Log.Info("wrong password")
	} else {
		d.Protected = false
	}
}

func (d *Drive) SetNewPassword(password string, newPassword string, isAdmin bool) {
	if !d.Protected {
		if password == d.Key {
			d.Key = newPassword
			d.Log.Info("password updated")
		} else {
			d.Log.Info("wrong password")
		}
	} else {
		if isAdmin {
			d.Log.Warn("admin password override!")
			d.Key = newPassword
			return
		}
		d.Log.Info(fmt.Sprintf("drive (id=%s) is protected. unlock with password.", d.ID))
	}
}

// ------- file management --------------------------------

// add file to a directory
func (d *Drive) AddFile(dirID string, file *File) error {
	if !d.Protected {
		if !d.HasRoot() {
			return fmt.Errorf("drive has no root directory")
		}
		if dirID == d.Root.ID {
			if err := d.Root.AddFile(file); err != nil {
				return err
			}
		} else {
			dir := d.GetDir(dirID)
			if dir == nil {
				return fmt.Errorf("dir (id=%s) not found", dirID)
			}
			if err := dir.AddFile(file); err != nil {
				return err
			}
		}
		d.UpdateDriveSize(file.GetSize())
	} else {
		d.Log.Info(fmt.Sprintf("drive (id=%s) is protected", d.ID))
	}
	return nil
}

// find a file. returns nil if not found.
func (d *Drive) GetFile(fileID string) *File {
	if !d.Protected {
		if !d.HasRoot() {
			d.Log.Error(fmt.Sprintf("drive (id=%s) has no root directory", d.ID))
			return nil
		}
		return d.Root.WalkF(fileID)
	} else {
		d.Log.Info(fmt.Sprintf("drive (id=%s) is protected", d.ID))
		return nil
	}
}

// get a slice of all files in the drive.
//
// returns an empty slice if the drive is protected, or if no files are found.
func (d *Drive) GetFiles() []*File {
	if !d.Protected {
		var (
			fm    = d.GetFilesMap()
			files = make([]*File, 0, len(fm))
		)
		for _, f := range fm {
			files = append(files, f)
		}
		return files
	} else {
		d.Log.Info(fmt.Sprintf("drive (id=%s) is protected", d.ID))
		return make([]*File, 0)
	}
}

// get a map of all available files for this user
func (d *Drive) GetFilesMap() map[string]*File {
	if !d.Protected {
		if !d.HasRoot() {
			d.Log.Error(fmt.Sprintf("drive (id=%s) has no root directory", d.ID))
			return nil
		}
		return d.Root.WalkFs()
	} else {
		d.Log.Info(fmt.Sprintf("drive (id=%s) is protected", d.ID))
	}
	return nil
}

// update files metadata in the drive
func (d *Drive) UpdateFile(dirID string, file *File) error {
	if !d.Protected {
		if d.Root.ID == dirID {
			if err := d.Root.PutFile(file); err != nil {
				return fmt.Errorf("failed to update file %s: %v", file.ID, err)
			}
		} else {
			dir := d.GetDir(dirID)
			if dir == nil {
				return fmt.Errorf("dir (id=%s) not found", dirID)
			}
			var origSize = file.GetSize()
			if err := dir.PutFile(file); err != nil {
				return err
			}
			var newSize = file.GetSize()
			d.UpdateDriveSize(origSize - newSize)
		}
	} else {
		d.Log.Info(fmt.Sprintf("drive (id=%s) is protected", d.ID))
	}
	return nil
}

// removes physical file from original location and
// updates internal data structures. use with caution!
func (d *Drive) RemoveFile(dirID string, file *File) error {
	if !d.Protected {
		if !d.HasRoot() {
			return fmt.Errorf("no root directory")
		}
		// if the driveID is this drive's root directory
		if dirID == d.Root.ID {
			if err := d.Root.RemoveFile(file.ID); err != nil {
				return err
			}
		} else {
			// check subdirectories
			dir := d.GetDir(dirID)
			if dir == nil {
				return fmt.Errorf("dir (id=%s) not found", dirID)
			}
			d.UpdateDriveSize(-file.GetSize())
			if err := dir.RemoveFile(file.ID); err != nil {
				return err
			}
			return nil
		}
	} else {
		d.Log.Info(fmt.Sprintf("drive (id=%s) is protected", d.ID))
	}
	return nil
}

// ------ directory management --------------------------------

func (d *Drive) addSubDir(dirID string, dir *Directory) error {
	// add sub directory to root if that's where it's supposed to be
	if dirID == d.Root.ID {
		if err := d.Root.AddSubDir(dir); err != nil {
			return err
		}
	} else {
		// otherwise attempt to retrive the directory we want to
		// add the subdirectory to
		dirs := d.Root.GetDirMap()
		if d, exists := dirs[dirID]; exists {
			return d.AddSubDir(dir)
		} else {
			return fmt.Errorf("drive has no directory with id=%s", dirID)
		}
	}
	return nil
}

// add a sub directory to a directory within the drive file system.
// does NOT create a physical sub directory! only adds metadata to the
// directory management system.
func (d *Drive) AddSubDir(dirID string, dir *Directory) error {
	if !d.Protected {
		if !d.HasRoot() {
			return fmt.Errorf("no root directory")
		}
		if err := d.addSubDir(dirID, dir); err != nil {
			return err
		}
		// Client is only concered about the directory size. server doesnt' care.
		if dir.LocalBackup {
			size, err := dir.GetSize()
			if err != nil {
				return err
			}
			d.UpdateDriveSize(size)
		}
	} else {
		d.Log.Info(fmt.Sprintf("drive (id=%s) is protected", d.ID))
	}
	return nil
}

// add subdirectories to the drives root directory
func (d *Drive) AddDirs(dirs []*Directory) error {
	if !d.Protected {
		if !d.HasRoot() {
			return fmt.Errorf("no root directory")
		}
		if err := d.Root.AddSubDirs(dirs); err != nil {
			return err
		}
		var total int64
		for _, dir := range dirs {
			size, err := dir.GetSize()
			if err != nil {
				return err
			}
			total += size
		}
		d.UpdateDriveSize(total)
	} else {
		d.Log.Info(fmt.Sprintf("drive (id=%s) is protected", d.ID))
	}
	return nil
}

// find a directory using its ID.
// returns nil if not found (or if drive has no root directory)
func (d *Drive) GetDir(dirID string) *Directory {
	if !d.Protected {
		if !d.HasRoot() {
			d.Log.Warn(fmt.Sprintf("drive (id=%s) has no root dir. cannot traverse.", d.ID))
			return nil
		}
		if d.Root.ID == dirID {
			return d.Root
		}
		return d.Root.WalkD(dirID)
	} else {
		d.Log.Info(fmt.Sprintf("drive (id=%s) is protected", d.ID))
	}
	return nil
}

// get a slice of all directories in the drive. returns nil if not found,
// or if the drive is protected.
func (d *Drive) GetDirs() []*Directory {
	if !d.Protected {
		var (
			dm   = d.GetDirsMap()
			dirs = make([]*Directory, 0, len(dm))
		)
		for _, dir := range dm {
			dirs = append(dirs, dir)
		}
		return dirs
	} else {
		d.Log.Info(fmt.Sprintf("drive (id=%s) is protected", d.ID))
	}
	return nil
}

// get a map of all directories for this user. returns nil
// if none are found or if the drive is protected.
func (d *Drive) GetDirsMap() map[string]*Directory {
	if !d.Protected {
		if !d.HasRoot() {
			d.Log.Warn(fmt.Sprintf("drive (id=%s) has no root dir. cannot traverse.", d.ID))
			return nil
		}
		return d.Root.WalkDs()
	} else {
		d.Log.Info(fmt.Sprintf("drive (id=%s) is protected", d.ID))
	}
	return nil
}

// remove a directory from the drive service. does not remove physical directories!
func (d *Drive) removeDir(dirID string) error {
	dir := d.GetDir(dirID)
	if dir != nil {
		if dir.Parent != nil {
			delete(dir.Parent.Dirs, dir.ID)
		}
		_ = d.Root.RemoveSubDir(dir.ID)
		d.Log.Info(fmt.Sprintf("directory (id=%s) removed", dirID))
	} else {
		d.Log.Info(fmt.Sprintf("directory (id=%s) not found", dirID))
	}
	return nil
}

// remove a directory entry from the sfs filesystem.
func (d *Drive) RemoveDir(dirID string) error {
	if !d.Protected {
		if !d.HasRoot() {
			return fmt.Errorf("no root directory")
		}
		if err := d.removeDir(dirID); err != nil {
			return err
		}
	} else {
		d.Log.Info(fmt.Sprintf("drive (id=%s) is protected", d.ID))
	}
	return nil
}

// remove directories from the drive root directory
func (d *Drive) RemoveDirs(dirs []*Directory) error {
	if !d.Protected {
		if !d.HasRoot() {
			return fmt.Errorf("no root directory")
		}
		if len(dirs) == 0 {
			return nil
		}
		for _, dir := range dirs {
			if err := d.RemoveDir(dir.ID); err != nil {
				return err
			}
		}
	} else {
		d.Log.Info(fmt.Sprintf("drive (id=%s) is protected", d.ID))
	}
	return nil
}

// update a directory within a drive.
func (d *Drive) UpdateDir(dirID string, updatedDir *Directory) error {
	if !d.Protected {
		if !d.HasRoot() {
			return fmt.Errorf("no root directory")
		}
		// get the parent of the updated directory & add
		parent := d.Root.WalkD(dirID)
		if parent == nil {
			return fmt.Errorf("dir %s not found", dirID)
		}
		if err := parent.PutSubDir(updatedDir); err != nil {
			return fmt.Errorf("failed to update dir %s: %v", parent.Name, err)
		}
	} else {
		d.Log.Info(fmt.Sprintf("drive (id=%s) is protected", d.ID))
	}
	return nil
}

// ----- cleanup --------------------------------

// removes all users files and directories from their drive
func (d *Drive) ClearDrive() error {
	if !d.Protected {
		if !d.HasRoot() {
			return fmt.Errorf("no root directory")
		}
		if err := d.Root.Clean(d.Root.Path); err != nil {
			return err
		}
	} else {
		d.Log.Info(fmt.Sprintf("drive (id=%s) is protected", d.ID))
	}
	return nil
}

// ---- sync operations --------------------------------

func (d *Drive) BuildSyncIdx() {
	files := d.GetFiles()
	d.SyncIndex = BuildSyncIndex(files, nil, d.SyncIndex)
}

func (d *Drive) BuildToUpdate() {
	if d.SyncIndex == nil {
		d.SyncIndex = NewSyncIndex(d.OwnerID)
	}
	d.SyncIndex = BuildToUpdate(d.GetFiles(), nil, d.SyncIndex)
}
