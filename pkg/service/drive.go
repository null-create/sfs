package service

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/sfs/pkg/auth"
)

// max size of a single drive (root directory) per user (1GB)
const MAX_SIZE float64 = 1e+9

/*
build a new privilaged drive directory for a client on the sfs server
with base state file info for user and drive json files

must be under ../svcroot/users/<username>

drives should have the following structure:

user/
|---root/    <---- user files & directories live here
|---state/
|   |---userID-d-m-y-hh-mm-ss.json
|   |---driveID-d-m-y-hh-mm-ss.json
*/
func AllocateDrive(name string, ownerID string, svcRoot string) (*Drive, error) {
	// new user service file paths
	userRoot := filepath.Join(svcRoot, "users", name)
	contentsRoot := filepath.Join(userRoot, "root")
	stateFileDir := filepath.Join(userRoot, "state")

	// make each directory
	dirs := []string{userRoot, contentsRoot, stateFileDir}
	for _, d := range dirs {
		if err := os.Mkdir(d, PERMS); err != nil {
			return nil, err
		}
	}

	// gen root and drive objects
	driveID := auth.NewUUID()
	rt := NewRootDirectory(name, ownerID, driveID, contentsRoot)
	drv := NewDrive(driveID, name, ownerID, userRoot, rt.ID, rt)
	return drv, nil
}

/*
"Drives" are just abstractions of a protrected root directory,
managed by Simple File Sync, containing backups of a user's files and other subdirectories
to facilitate synchronization across multiple devices.

It's basically just a directory containing some metadata about a users
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
	TotalSize float64 `json:"total_size"`
	UsedSpace float64 `json:"used_space"`
	FreeSpace float64 `json:"free_space"`

	// Security stuff
	Protected bool   `json:"protected"`
	Key       string `json:"key"`
	AuthType  string `json:"auth_type"`

	// location of the drive on physical server filesystem
	// i.e., ...sfs/root/users/this-drive
	DriveRoot string `json:"drive_root"`

	// Flag for whether Populate() has been called
	// with the drive's root directory. If so then the
	// drive's root directory will have its internal data structures
	// loaded and will make other calls to the users file contents possible.
	IsLoaded bool `json:"is_loaded"`

	// User's root directory & sync index
	RootID    string     `json:"root_id"`
	Root      *Directory `json:"-"` // ignored to avoid json cycle errors
	SyncIndex *SyncIndex `json:"sync_index"`
}

func check(driveID string, ownerName string, ownerID string, rootPath string, root *Directory) bool {
	if driveID == "" || ownerName == "" || ownerID == "" || rootPath == "" || root == nil {
		log.Printf("[ERROR] invalid drive parameters. none can be empty!")
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
		ID:        driveID,
		OwnerName: ownerName,
		OwnerID:   ownerID,
		TotalSize: MAX_SIZE,
		UsedSpace: 0,
		FreeSpace: MAX_SIZE,
		Protected: false,
		Key:       "default",
		DriveRoot: rootPath,
		RootID:    rootID,
		Root:      root,
	}
}

func (d *Drive) RemainingSize() float64 {
	return d.TotalSize - d.UsedSpace
}

func (d *Drive) Isloaded() bool {
	return d.IsLoaded
}

// save drive state to JSON format
func (d *Drive) ToJSON() ([]byte, error) {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return nil, err
	}
	return data, nil
}

// save drive state to JSON format in current directory.
func (d *Drive) SaveState() error {
	data, err := d.ToJSON()
	if err != nil {
		return err
	}
	fn := fmt.Sprintf("user-%s-.json", time.Now().UTC().Format("2006-01-02T15-04-05"))
	fp := filepath.Join(d.DriveRoot, "state", fn)
	return os.WriteFile(fp, data, PERMS)
}

func (d *Drive) GetOwnerID() (string, error) {
	if d.OwnerID == "" {
		return "", fmt.Errorf("drives should have an associated owner ID")
	}
	return d.OwnerID, nil
}

// ------- security --------------------------------

func (d *Drive) Lock(password string) {
	if password != d.Key {
		log.Printf("[DEBUG] wrong password: %v", password)
	} else {
		d.Protected = true
	}
}

func (d *Drive) Unlock(password string) {
	if password != d.Key {
		log.Printf("[DEBUG] wrong password: %v", password)
	} else {
		d.Protected = false
	}
}

func (d *Drive) SetNewPassword(password string, newPassword string, isAdmin bool) {
	if !d.Protected {
		if password == d.Key {
			d.Key = newPassword
			log.Printf("[DEBUG] password updated")
		} else {
			log.Print("[DEBUG] wrong password")
		}
	} else {
		if isAdmin {
			log.Print("[DEBUG] admin password override")
			d.Key = newPassword
			return
		}
		log.Print("[DEBUG] drive protected. unlock with password.")
	}
}

// ------- file management --------------------------------

// add file to a directory
func (d *Drive) AddFile(dirID string, file *File) error {
	if !d.Protected {
		if dirID == d.Root.ID {
			d.Root.AddFile(file)
		} else {
			dir := d.GetDir(dirID)
			if dir == nil {
				return fmt.Errorf("dir (id=%s) not found", dirID)
			}
			dir.AddFile(file)
		}
		d.UsedSpace += float64(file.Size())
	} else {
		log.Printf("[DEBUG] drive (id=%s) is protected", d.ID)
	}
	return nil
}

// find a file. returns nil if not found.
func (d *Drive) GetFile(fileID string) *File {
	if !d.Protected {
		return d.Root.WalkF(fileID)
	} else {
		log.Printf("[DEBUG] drive (id=%s) is protected", d.ID)
	}
	return nil
}

// get a map of all available files for this user
func (d *Drive) GetFiles() map[string]*File {
	if !d.Protected {
		return d.Root.WalkFs()
	} else {
		log.Printf("[DEBUG] drive (id=%s) is protected", d.ID)
	}
	return nil
}

// update a file
func (d *Drive) UpdateFile(dirID string, file *File, data []byte) error {
	if !d.Protected {
		if d.Root.ID == dirID {
			if err := d.Root.UpdateFile(file, data); err != nil {
				return fmt.Errorf("failed to update file %s: %v", file.ID, err)
			}
		} else {
			dir := d.GetDir(dirID)
			if dir == nil {
				return fmt.Errorf("dir (id=%s) not found", dirID)
			}
			if err := dir.UpdateFile(file, data); err != nil {
				return err
			}
			// TODO: get the difference between old and new file sizes
			// and adjust the drives used space value accordingly.
		}
	} else {
		log.Printf("[INFO] drive is protected")
	}
	return nil
}

// remove file from a directory. removes physical file and
// updates internal data structures.
func (d *Drive) RemoveFile(dirID string, file *File) error {
	if !d.Protected {
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
			if err := dir.RemoveFile(file.ID); err != nil {
				return err
			}
			d.UsedSpace -= float64(file.Size())
			return nil
		}
	} else {
		log.Printf("[INFO] drive is protected")
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
		dirs := d.Root.GetSubDirs()
		if d, exists := dirs[dirID]; exists {
			return d.AddSubDir(dir)
		} else {
			return fmt.Errorf("drive has no directory with id=%s", dirID)
		}
	}
	return nil
}

// add a sub directory to a directory within the drive file system.
// creates a physical sub directory at the path assigned within
// the directory parameter.
func (d *Drive) AddSubDir(dirID string, dir *Directory) error {
	if !d.Protected {
		if err := d.addSubDir(dirID, dir); err != nil {
			return err
		}
	} else {
		log.Printf("[DEBUG] drive (id=%s) is protected", d.ID)
	}
	return nil
}

// add subdirectories to the drives root directory
func (d *Drive) AddDirs(dirs []*Directory) error {
	if !d.Protected {
		if err := d.Root.AddSubDirs(dirs); err != nil {
			return err
		}
	} else {
		log.Printf("[DEBUG] drive (id=%s) is protected", d.ID)
	}
	return nil
}

// find a directory. returns nil if not found (or if drive has no root directory)
func (d *Drive) GetDir(dirID string) *Directory {
	if !d.Protected {
		if d.Root == nil {
			log.Printf("[WARNING] drive (id=%s) has no root dir. cant traverse.", d.ID)
			return nil
		}
		if d.Root.ID == dirID {
			return d.Root
		}
		return d.Root.WalkD(dirID)
	} else {
		log.Printf("[DEBUG] drive (id=%s) is protected", d.ID)
	}
	return nil
}

// get a map of all directories for this user
func (d *Drive) GetDirs() map[string]*Directory {
	if !d.Protected {
		return d.Root.WalkDs()
	} else {
		log.Printf("[DEBUG] drive (id=%s) is protected", d.ID)
	}
	return nil
}

func (d *Drive) removeDir(dirID string) error {
	dir := d.GetDir(dirID)
	if dir != nil {
		if err := os.RemoveAll(dir.Path); err != nil {
			return err
		}
		// this should only really apply to any non-root directory
		if dir.Parent != nil {
			delete(dir.Parent.Dirs, dir.ID)
		}
		log.Printf("[DEBUG] directory (id=%s) removed", dirID)
	} else {
		log.Printf("[DEBUG] directory (id=%s) not found", dirID)
	}
	return nil
}

// removed a directory from the drive.
// removes physical drive and all its children,
// as well as deletes the directory entry from the
// sfs filesystem.
func (d *Drive) RemoveDir(dirID string) error {
	if !d.Protected {
		if err := d.removeDir(dirID); err != nil {
			return err
		}
	} else {
		log.Printf("[DEBUG] drive (id=%s) is protected", d.ID)
	}
	return nil
}

// remove directories from the drive root directory
func (d *Drive) RemoveDirs(dirs []*Directory) error {
	if !d.Protected {
		if len(dirs) == 0 {
			log.Print("[INFO] no subdirectories to remove")
			return nil
		}
		for _, dir := range dirs {
			if err := d.RemoveDir(dir.ID); err != nil {
				return err
			}
		}
	} else {
		log.Printf("[DEBUG] drive (id=%s) is protected", d.ID)
	}
	return nil
}

// ----- cleanup --------------------------------

// removes all users files and directories from their drive
func (d *Drive) ClearDrive() error {
	if !d.Protected {
		if err := d.Root.Clean(d.Root.Path); err != nil {
			return err
		}
	} else {
		log.Printf("[DEBUG] drive protected")
	}
	return nil
}

// ---- sync operations --------------------------------

func (d *Drive) BuildSyncIdx() error {
	idx := BuildSyncIndex(d.Root)
	if idx == nil {
		return fmt.Errorf("unable to build sync index")
	}
	d.SyncIndex = idx
	return nil
}

func (d *Drive) BuildToUpdate() error {
	if d.SyncIndex == nil {
		return fmt.Errorf("no sync index to build from")
	}
	d.SyncIndex = BuildToUpdate(d.Root, d.SyncIndex)
	if len(d.SyncIndex.ToUpdate) == 0 {
		return fmt.Errorf("no files matched for syncing")
	}
	return nil
}
