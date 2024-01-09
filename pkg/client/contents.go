package client

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	svc "github.com/sfs/pkg/service"
)

// ----- files --------------------------------------

// list all local files managed by the sfs service.
// does not check database
func (c *Client) ListLocalFiles() {
	files := c.Drive.GetFiles()
	for _, f := range files {
		fmt.Print(fmt.Sprintf("id: %s\n name: %s\n loc: %s", f.ID, f.Name, f.ClientPath))
	}
}

// retrieve a file. returns nil if the file is not found.
func (c *Client) GetFile(fileID string) (*svc.File, error) {
	file := c.Drive.GetFile(fileID)
	if file == nil {
		// try database before giving up.
		file, err := c.Db.GetFile(fileID)
		if err != nil {
			return nil, err
		}
		if file == nil {
			return nil, fmt.Errorf("file %s not found", fileID)
		}
		// add this since we didn't have it before
		if err := c.Drive.AddFile(file.DirID, file); err != nil {
			return nil, fmt.Errorf("failed to add file %s: %v", file.DirID, err)
		}
		if err := c.SaveState(); err != nil {
			return nil, fmt.Errorf("failed to save state: %v", err)
		}
	}
	return file, nil
}

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
		drive.Root = c.Populate(root)
		if !drive.IsIndexed() {
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
			return nil, fmt.Errorf("failed to add file to database: %v", err)
		}
	}
	dirs := root.WalkDs()
	for _, d := range dirs {
		if err := c.Db.AddDir(d); err != nil {
			return nil, fmt.Errorf("failed to add directory to database: %v", err)
		}
	}
	// add root directory itself
	if err := c.Db.AddDir(root); err != nil {
		return nil, fmt.Errorf("failed to add root to database: %v", err)
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

// recursively descends the drive's directory tree and compares what it
// finds to what is in the database, adding new items as it goes. generates
// a new root directory object and attaches it to the drive.
func (c *Client) RefreshDrive() error {
	// refresh root against the database and create a new root object
	newRoot := c.refreshDrive(c.Drive.Root)

	// clear old contents from memory then add new root
	c.Drive.Root.Clear(c.Drive.Root.Key)
	c.Drive.Root = newRoot

	if err := c.SaveState(); err != nil {
		return fmt.Errorf("failed to save state file: %v", err)
	}
	return nil
}

// descends users directory tree and compares what it finds
// to what is in the database. if a new file or directory is found
// along the way it will be added to the database and new objects
// will created for them.
func (c *Client) refreshDrive(dir *svc.Directory) *svc.Directory {
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
			// new directory
			if subDir == nil {
				subDir = svc.NewDirectory(item.Name(), dir.OwnerID, dir.DriveID, entryPath)
				if err := c.Db.AddDir(subDir); err != nil {
					log.Printf("[ERROR] could not add directory (%s) to db: %v", item.Name(), err)
					continue
				}
				subDir = c.refreshDrive(subDir)
				dir.AddSubDir(subDir)
			}
		} else {
			file, err := c.Db.GetFileByName(item.Name())
			if err != nil {
				log.Printf("[ERROR] could not get file (%s) from db: %v", item.Name(), err)
				continue
			}
			// new file
			if file == nil {
				newFile := svc.NewFile(item.Name(), dir.DriveID, dir.OwnerID, filepath.Join(item.Name(), dir.Path))
				if err := c.Db.AddFile(newFile); err != nil {
					log.Printf("[ERROR] could not add file (%s) to db: %v", item.Name(), err)
					continue // TEMP until there's a better way to handle this error
				}
				dir.AddFile(newFile)
			}
		}
	}
	return dir
}
