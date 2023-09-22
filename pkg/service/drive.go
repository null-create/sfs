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
	TotalSize float64 `json:"totalSize"`
	UsedSpace float64 `json:"usedSpace"`
	FreeSpace float64 `json:"freeSpace"`

	// Security stuff
	Protected bool   `json:"protected"`
	Key       string `json:"key"`
	AuthType  string `json:"authType"`

	// location of the drive on physical server filesystem
	DriveRoot string `json:"driveRoot"`

	// User's root directory
	Root *Directory `json:"root"`
}

func check(id string, name string, owner string, rootPath string, root *Directory) bool {
	if id == "" || name == "" || owner == "" || rootPath == "" || root == nil {
		log.Printf("[ERROR] invalid drive parameters. none can be empty!")
		return false
	}
	return true
}

// creates a new drive service for a user. does not create new physical files
func NewDrive(id string, name string, owner string, rootPath string, root *Directory) *Drive {
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

// TODO: figure out where to place the file
// based on the supplied file object (containing the path)
func (d *Drive) AddFile(f *File) error {
	return nil
}

// find a file
func (d *Drive) GetFile(fileID string) *File {
	if !d.Protected {
		if len(d.Root.Files) != 0 {
			if f, ok := d.Root.Files[fileID]; ok {
				return f
			}
		}
		if len(d.Root.Dirs) == 0 {
			return nil
		}
		return d.Root.WalkF(fileID)
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
	}
	return nil
}
