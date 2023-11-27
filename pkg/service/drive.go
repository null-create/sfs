package service

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// max size of a single drive (root directory) per user (1GB)
const MAX_SIZE float64 = 1e+9

/*
"Drives" are just abstractions of a protrected root directory,
managed by Simple File Sync, containing backups of a user's files and other subdirectories
to facilitate synchronization across multiple devices.

It's basically just a directory containing some metadata about a users
drive size, state, and some security configurations, which itself
contains the user's "root" directory. Its this "root" directory where
all the users files live, in whatever arragement they end up using.

user/
|---meta/
|   |---user.json
|   |---drive.json
|---root/    <---- the "drive." user's files & directories live here
|---state/
|   |---userID-d-m-y-hh-mm-ss.json

Drives may be realized as a filesystem on a user's current desktop,
laptop, dedicated hardrive within a desktop, or separate server.
*/
type Drive struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Owner string `json:"owner:"`

	// all three measured in Kb or Mb
	TotalSize float64 `json:"total_size"`
	UsedSpace float64 `json:"used_space"`
	FreeSpace float64 `json:"free_space"`

	// Security stuff
	Protected bool   `json:"protected"`
	Key       string `json:"key"`
	AuthType  string `json:"auth_type"`

	// location of the drive on physical server filesystem
	DriveRoot string `json:"drive_root"`

	// User's root directory & sync index
	RootID    string     `json:"root_id"`
	Root      *Directory `json:"root"`
	SyncIndex *SyncIndex `json:"sync_index"`
}

func check(id string, name string, owner string, rootPath string, root *Directory) bool {
	if id == "" || name == "" || owner == "" || rootPath == "" || root == nil {
		log.Printf("[ERROR] invalid drive parameters. none can be empty!")
		return false
	}
	return true
}

// creates a new drive service for a user. does not create new physical files
func NewDrive(id string, name string, owner string, rootPath string, rootID string, root *Directory) *Drive {
	if !check(id, name, owner, rootPath, root) {
		return nil
	}
	return &Drive{
		ID:    id,
		Name:  name,
		Owner: owner,

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

// save drive state to JSON format
func (d *Drive) ToJSON() ([]byte, error) {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *Drive) SaveState() error {
	data, err := d.ToJSON()
	if err != nil {
		return err
	}
	fn := fmt.Sprintf("user-%s-.json", time.Now().UTC().Format("2006-01-02T15-04-05"))
	fp := filepath.Join(d.DriveRoot, fn)
	return os.WriteFile(fp, data, 0644)
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
		dir := d.GetDir(dirID)
		if dir == nil {
			return fmt.Errorf("dir (id=%s) not found", dirID)
		}
		dir.AddFile(file)
		d.Root.Size += float64(file.Size())
	} else {
		log.Printf("[DEBUG] drive (id=%s) is protected", d.ID)
	}
	return nil
}

// find a file
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
	dir := d.GetDir(dirID)
	if dir == nil {
		return fmt.Errorf("dir (id=%s) not found", dirID)
	}
	if err := dir.UpdateFile(file, data); err != nil {
		return err
	}
	return nil
}

// remove file from a directory
func (d *Drive) RemoveFile(dirID string, file *File) error {
	dir := d.GetDir(dirID)
	if dir == nil {
		return fmt.Errorf("dir (id=%s) not found", dirID)
	}
	if err := dir.RemoveFile(file.ID); err != nil {
		return err
	}
	d.Root.Size -= float64(file.Size())
	return nil
}

// ------ directory management --------------------------------

func (d *Drive) addDir(dir *Directory) error {
	if err := d.Root.AddSubDir(dir); err != nil {
		return err
	}
	return nil
}

// add a directory. currently defaults to drive.Root
func (d *Drive) AddDir(dir *Directory) error {
	if !d.Protected {
		if _, exists := d.Root.Dirs[dir.ID]; !exists {
			if err := d.addDir(dir); err != nil {
				return err
			}
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

// find a directory
func (d *Drive) GetDir(dirID string) *Directory {
	if !d.Protected {
		if d.Root.ID == dirID {
			return d.Root
		}
		return d.Root.WalkD(dirID)
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
		parent := dir.Parent
		if parent != nil {
			delete(parent.Dirs, dir.ID)
		}
		log.Printf("[DEBUG] directory (id=%s) removed", dirID)
	} else {
		log.Printf("[DEBUG] directory (id=%s) not found", dirID)
	}
	return nil
}

// remove a directory
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
