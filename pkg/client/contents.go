package client

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	svc "github.com/sfs/pkg/service"
)

// ----- files --------------------------------------

// add a file to a specified directory
func (c *Client) AddFile(dirID string, file *svc.File) error {
	return c.Drive.AddFile(dirID, file)
}

// add a series of files to a specified directory
func (c *Client) AddFiles(dirID string, files []*svc.File) error {
	if len(files) == 0 {
		return fmt.Errorf("no files to add")
	}
	for _, file := range files {
		if err := c.Drive.AddFile(dirID, file); err != nil {
			return err
		}
	}
	return nil
}

// update a file in a specied directory
func (c *Client) UpdateFile(dirID string, fileID string, data []byte) error {
	file := c.Drive.GetFile(fileID)
	if file == nil {
		return fmt.Errorf("no file ID %s found", fileID)
	}
	if len(data) == 0 {
		return fmt.Errorf("no data received")
	}
	return c.Drive.UpdateFile(dirID, file, data)
}

// remove a file in a specied directory
func (c *Client) RemoveFile(dirID string, file *svc.File) error {
	return c.Drive.RemoveFile(dirID, file)
}

// remove files from a specied directory
func (c *Client) RemoveFiles(dirID string, fileIDs []string) error {
	if len(fileIDs) == 0 {
		return fmt.Errorf("no files to remove")
	}
	for _, fileID := range fileIDs {
		file := c.Drive.GetFile(fileID)
		if file == nil { // didn't find the file. ignore.
			continue
		}
		if err := c.Drive.RemoveFile(dirID, file); err != nil {
			return err
		}
	}
	return nil
}

// ----- directories --------------------------------

func (c *Client) AddDir(dirID string, dir *svc.Directory) error {
	return c.Drive.AddSubDir(dirID, dir)
}

func (c *Client) AddDirs(dirs []*svc.Directory) error {
	return c.Drive.AddDirs(dirs)
}

func (c *Client) RemoveDir(dirID string) error {
	return c.Drive.RemoveDir(dirID)
}

func (c *Client) RemoveDirs(dirs []*svc.Directory) error {
	return c.Drive.RemoveDirs(dirs)
}

// ----- drive --------------------------------

// Populates drive's root directory with and all the users
// subdirectories and files from the database into memory.
// Should be followed by a call to Drive.Root.Clear() (not clean!)
// to clear the drive's internal data structures after loading.
func (c *Client) LoadDrive(driveID string) *svc.Drive {
	c.Drive.Root = c.Populate(c.Drive.Root)
	c.Drive.IsLoaded = true
	return c.Drive
}

// retrieves drive with root directory attached, but unpopulated.
// also updates drive sync index.
func (c *Client) GetDrive(driveID string) (*svc.Drive, error) {
	if c.Drive == nil {
		log.Print("[WARNING] drive not found! attempting to load...")
		// find drive by id
		drive, err := c.Db.GetDrive(driveID)
		if err != nil {
			return nil, err
		}
		// get root directory for the drive and create a sync index if necessary
		root, err := c.Db.GetDirectory(drive.RootID)
		if err != nil {
			return nil, err
		}
		drive.Root = root
		if drive.SyncIndex == nil {
			drive.SyncIndex = drive.Root.WalkS(svc.NewSyncIndex(drive.OwnerID))
		}
		c.Drive = drive
	}
	return c.Drive, nil
}

// save drive state to DB
func (c *Client) SaveDrive(drv *svc.Drive) error {
	if err := c.Db.UpdateDrive(drv); err != nil {
		return fmt.Errorf("failed to update drive in database: %v", err)
	}
	if err := drv.SaveState(); err != nil {
		return fmt.Errorf("failed to save drive state: %v", err)
	}
	return nil
}

// discover populates the given root directory with the users file and
// sub directories, updates the database as it does so, and returns
// the the directory object when finished, or if there was an error.
//
// this should ideally be used for starting a new sfs service in a
// users root directly that already has files and/or subdirectories.
func (c *Client) Discover(root *svc.Directory) (*svc.Directory, error) {
	// traverse users SFS file system and populate internal structures
	root = root.Walk()
	// send everything to the database
	files := root.WalkFs()
	for _, file := range files {
		if err := c.Db.AddFile(file); err != nil {
			return nil, fmt.Errorf("failed to add file to database: ", err)
		}
	}
	dirs := root.WalkDs()
	for _, d := range dirs {
		if err := c.Db.AddDir(d); err != nil {
			return nil, fmt.Errorf("failed to add directory to database: ", err)
		}
	}
	// add root directory itself
	if err := c.Db.AddDir(root); err != nil {
		return nil, fmt.Errorf("failed to add root to database: ", err)
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
func (c *Client) Populate(root *svc.Directory) *svc.Directory {
	if root.Path == "" {
		log.Print("[WARNING] can't traverse directory without a path")
		return nil
	}
	if root.IsNil() {
		log.Printf(
			"[WARNING] can't traverse directory with emptyr or nil maps: \nfiles=%v dirs=%v",
			root.Files, root.Dirs,
		)
		return nil
	}
	return c.populate(root)
}

func (c *Client) populate(dir *svc.Directory) *svc.Directory {
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
			subDir, err := c.Db.GetDirectoryByName(item.Name())
			if err != nil {
				log.Printf("[ERROR] could not get directory from db: %v \nerr: %v", item.Name(), err)
				continue
			}
			if subDir == nil {
				continue
			}
			subDir.Parent = dir
			subDir = c.populate(subDir)
			if err := dir.AddSubDir(subDir); err != nil {
				log.Printf("[ERROR] could not add directory: %v", err)
				continue
			}
		} else { // add file
			file, err := c.Db.GetFileByName(item.Name())
			if err != nil {
				log.Printf("[ERROR] could not get file (%s) from db: %v", item.Name(), err)
				continue
			}
			if file == nil {
				continue
			}
			dir.AddFile(file)
		}
	}
	return dir
}
