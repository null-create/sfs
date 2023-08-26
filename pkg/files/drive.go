package files

import (
	"log"
)

// max size of a single drive (root directory) per user (1GB)
const MAX_SIZE float64 = 1e+9

/*
"Drives" are just abstractions of a protrected root directory,
managed by Nimbus, containing backups of a user's files and other subdirectories
to facilitate synchronization across multiple devices.

It's basically just a directory containing some metadata about a users
drive size, state, and some security configurations, which itself
contains the user's "root" directory. Its this "root" directory where
all the users files live, in whatever arragement they end up using.

drive/
|----root/
|----user-info.json
|----drive-info.json
|----credentials.json

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
	Root *Directory
}

func check(id string, name string, owner string, rootPath string, root *Directory) bool {
	if id == "" || name == "" || owner == "" || rootPath == "" || root == nil {
		log.Printf("[ERROR] invalid drive parameters. none can be empty!")
		return false
	}
	return true
}

// Creates a new Drive service for a user.
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

func (d *Drive) GetRoot() *Directory {
	if d.Root == nil {
		log.Printf("[WARNING] no root directory assigned to this drive!")
		return nil
	}
	return d.Root
}

// --------- security --------------------------------

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

// --------- meta data --------------------------------

func (d *Drive) DriveSize() float64 {
	if len(d.Root.Dirs) == 0 {
		return 0.0
	}
	var total float64
	for _, dir := range d.Root.Dirs {
		size, err := dir.DirSize()
		if err != nil {
			log.Fatalf("[ERR] dir size error: %v", err)
		}
		total += size
	}
	return total
}
