package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	svc "github.com/sfs/pkg/service"
)

// ------ service --------------------------------

// add a file or directory to the local SFS service
// does not add the file to the SFS server.
func (c *Client) AddItem(itemPath string) error {
	item, err := os.Stat(itemPath)
	if err != nil {
		return err
	}
	if item.IsDir() {
		if err := c.AddDir(itemPath); err != nil {
			return err
		}
	} else {
		if err := c.AddFile(itemPath); err != nil {
			return err
		}
	}
	return nil
}

// remove an item from the local SFS service.
// does not remove the item from the server.
func (c *Client) RemoveItem(itemPath string) error {
	item, err := os.Stat(itemPath)
	if err != nil {
		return err
	}
	if item.IsDir() {
		dir, err := c.Db.GetDirectoryByPath(itemPath)
		if err != nil {
			return err
		}
		if err := c.RemoveDir(dir); err != nil {
			return err
		}
	} else {
		file, err := c.Db.GetFileByPath(itemPath)
		if err != nil {
			return nil
		}
		if err := c.RemoveFile(file); err != nil {
			return err
		}
	}
	return nil
}

// ----- files --------------------------------------

func (c *Client) Exists(path string) bool {
	if _, err := os.Stat(path); err != nil && errors.Is(err, os.ErrNotExist) {
		return false
	} else if err != nil {
		c.log.Error(fmt.Sprintf("failed to retrieve stat for: %s - %v", path, err))
		return false
	}
	return true
}

// list all local files managed by the sfs service.
// does not check database.
func (c *Client) ListLocalFiles() {
	files := c.Drive.GetFiles()
	for _, f := range files {
		output := fmt.Sprintf("id: %s\nname: %s\nloc: %s\n\n", f.ID, f.Name, f.ClientPath)
		fmt.Print(output)
	}
}

// list all files managed by the local sfs database
func (c *Client) ListLocalFilesDB() error {
	files, err := c.Db.GetUsersFiles(c.UserID)
	if err != nil {
		return err
	}
	for _, f := range files {
		fmt.Printf("id: %s\nname: %s\nloc: %s\n\n", f.ID, f.Name, f.ClientPath)
	}
	return nil
}

// list all files known to the remote SFS server
func (c *Client) ListRemoteFiles() error {
	req, err := c.GetAllFilesRequest(c.User)
	if err != nil {
		return err
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		c.dump(resp, true)
		return nil
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return err
	}
	fmt.Print(buf.String()) // TODO: fancier output
	return nil
}

// retrieve a local file using its ID. returns nil if the file is not found.
func (c *Client) GetFileByID(fileID string) (*svc.File, error) {
	file := c.Drive.GetFile(fileID)
	if file == nil {
		// try database before giving up.
		file, err := c.Db.GetFileByID(fileID)
		if err != nil {
			return nil, err
		}
		if file == nil {
			return nil, fmt.Errorf("file (id=%s) not found", fileID)
		}
		// add this since we didn't have it before
		if err := c.Drive.AddFile(file.DirID, file); err != nil {
			return nil, fmt.Errorf("failed to add file %s: %v", file.DirID, err)
		}
	}
	return file, nil
}

// check db using a given file path. returns nil if not found.
func (c *Client) GetFileByPath(path string) (*svc.File, error) {
	file, err := c.Db.GetFileByPath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file by path: %v", err)
	} else if file == nil {
		return nil, fmt.Errorf("%s not found", filepath.Base(path))
	}
	return file, nil
}

// retrieve a file from the database by searching with its name.
// returns nil if not found.
func (c *Client) GetFileByName(name string) (*svc.File, error) {
	file, err := c.Db.GetFileByName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get file from database: %v", err)
	}
	if file == nil {
		return nil, fmt.Errorf("file %s not found", name)
	}
	return file, nil
}

// add a file to the service using its file path.
// should check for whether the directory it resides in is
// monitored by SFS -- though not contingent on it!
//
// SFS should be able to monitor files outside of the designated root directory.
// if we add a file this way then we should automatically make a backup of it
// in the SFS root directory with each detected change.

// add a file to the client-side service. does not push file to server.
func (c *Client) AddFile(filePath string) error {
	// see if we already have this file in the system
	file, err := c.Db.GetFileByPath(filePath)
	if err != nil {
		return err
	}
	if file != nil {
		return fmt.Errorf("file already exists")
	}

	// create new file object
	newFile := svc.NewFile(filepath.Base(filePath), c.DriveID, c.UserID, filePath)

	// see if we already have the file's parent directory in the file system
	nfDir := filepath.Dir(filePath)
	dir, err := c.GetDirByPath(nfDir)
	if err != nil && strings.Contains(err.Error(), "does not exist") {
		// if the parent directory to this file doesn't exist in the file system,
		// then add it to the SFS root.
		newFile.DirID = c.Drive.RootID
	} else if err != nil {
		return err
	} else {
		// directory already exists. add file to this directory.
		newFile.DirID = dir.ID
	}
	// add file to sfs system
	if err := c.Drive.AddFile(newFile.DirID, newFile); err != nil {
		return err
	}
	if err := c.Db.AddFile(newFile); err != nil {
		return err
	}
	// add file to monitoring system
	if err := c.WatchItem(filePath); err != nil {
		return err
	}
	// push metadata to server if autosync is enabled
	if c.autoSync() {
		req, err := c.NewFileRequest(newFile)
		if err != nil {
			return err
		}
		resp, err := c.Client.Do(req)
		if err != nil {
			return err
		}
		c.dump(resp, true)
	}
	return nil
}

// add a new file to a specified directory using a directory ID.
// adds file to database and monitoring services.
func (c *Client) AddFileWithID(dirID string, file *svc.File) error {
	if err := c.Drive.AddFile(dirID, file); err != nil {
		return err
	}
	if err := c.Db.AddFile(file); err != nil {
		return err
	}
	if err := c.WatchItem(file.ClientPath); err != nil {
		return err
	}
	// push metadata to server if autosync is enabled
	if c.autoSync() {
		req, err := c.NewFileRequest(file)
		if err != nil {
			return err
		}
		resp, err := c.Client.Do(req)
		if err != nil {
			return err
		}
		c.dump(resp, true)
	}
	return nil
}

// update file contents in a specied directory
func (c *Client) ModifyFile(dirID string, fileID string, data []byte) error {
	file := c.Drive.GetFile(fileID)
	if file == nil {
		return fmt.Errorf("no file (id=%s) found", fileID)
	}
	if len(data) == 0 {
		return fmt.Errorf("no data received")
	}
	if err := c.Drive.ModifyFile(dirID, file, data); err != nil {
		return err
	}
	if err := c.Db.UpdateFile(file); err != nil {
		return err
	}
	return nil
}

// update file metadata in the service instance
func (c *Client) UpdateFile(updatedFile *svc.File) error {
	oldFile := c.Drive.GetFile(updatedFile.ID)
	if oldFile == nil {
		return fmt.Errorf("file (id=%s) not found", updatedFile.ID)
	}
	if err := c.Drive.UpdateFile(oldFile.DirID, updatedFile); err != nil {
		return fmt.Errorf("failed to update file (id=%s): %v", updatedFile.ID, err)
	}
	if err := c.Db.UpdateFile(updatedFile); err != nil {
		return fmt.Errorf("failed to update file (id=%s) in database: %v", updatedFile.ID, err)
	}
	return nil
}

// remove a file in a specied directory.
func (c *Client) RemoveFile(file *svc.File) error {
	// stop monitoring the file
	c.Monitor.StopWatching(file.Path)
	// we're implementing "soft" deletes here. if a user wants to
	// actually Delete a file, we can implement another function for that later.
	// remove from drive and database
	if err := file.Copy(filepath.Join(c.RecycleBin, file.Name)); err != nil {
		return fmt.Errorf("failed to copy file to recyle directory: %v", err)
	}
	// remove physical file from original location
	if err := c.Drive.RemoveFile(file.DirID, file); err != nil {
		return err
	}
	if err := c.Db.RemoveFile(file.ID); err != nil {
		return err
	}
	return nil
}

// move a file from one directory to another. set keepOrig to true
// to keep a copy in the original local.
func (c *Client) MoveFile(destDirID string, file *svc.File, keepOrig bool) error {
	origDir := c.Drive.GetDir(file.DirID)
	if origDir == nil {
		return fmt.Errorf("original directory for file not found. dir id=%s", file.DirID)
	}
	destDir := c.Drive.GetDir(destDirID)
	if destDir == nil {
		return fmt.Errorf("destination directory (id=%s) not found", destDirID)
	}
	// add file object to destination directory
	if err := destDir.AddFile(file); err != nil {
		return err
	}
	// copy physical file
	if err := file.Copy(filepath.Join(destDir.Path, file.Name)); err != nil {
		return err
	}
	if !keepOrig {
		// remove from origial file (also deletes original physical file)
		if err := origDir.RemoveFile(file.ID); err != nil {
			return err
		}
	}
	// update dbs
	if err := c.Db.UpdateDir(origDir); err != nil {
		return err
	}
	if err := c.Db.UpdateDir(destDir); err != nil {
		return err
	}
	if err := c.Db.UpdateFile(file); err != nil {
		return err
	}
	return nil
}

// ----- directories --------------------------------

// list all directories managed by the SFS client
func (c *Client) ListLocalDirsDB() error {
	dirs, err := c.Db.GetUsersDirectories(c.UserID)
	if err != nil {
		return err
	}
	if len(dirs) == 0 {
		fmt.Print("no directories found")
		return nil
	}
	for _, dir := range dirs {
		var output = fmt.Sprintf(
			"name: %s\n id: %v\n loc: %s\n\n",
			dir.Name, dir.ID, dir.ClientPath,
		)
		fmt.Print(output)
	}
	return nil
}

// add dir to client service instance
func (c *Client) AddDirWithID(dirID string, dir *svc.Directory) error {
	if err := c.Drive.AddSubDir(dirID, dir); err != nil {
		return fmt.Errorf("failed to add directory: %v", err)
	}
	if err := c.Db.AddDir(dir); err != nil {
		// remove dir. we only want directories we have a record for.
		if remErr := os.Remove(dir.Path); remErr != nil {
			c.log.Warn(fmt.Sprintf("failed to remove directory: %v", remErr))
		}
		return err
	}
	if err := c.WatchItem(dir.Path); err != nil {
		return err
	}
	// push metadata to server if autosync is enabled
	if c.autoSync() {
		req, err := c.NewDirectoryRequest(dir)
		if err != nil {
			return err
		}
		resp, err := c.Client.Do(req)
		if err != nil {
			return err
		}
		c.dump(resp, true)
	}
	return nil
}

// add a new directory to the sfs file system. if its parent directory
// is already known, it will be added to there, otherwise will automatically
// be placed under root.
func (c *Client) AddDir(dirPath string) error {
	dir, err := c.Db.GetDirectoryByPath(dirPath)
	if err != nil {
		return err
	}
	if dir != nil {
		return fmt.Errorf("%s already exists in sfs system", filepath.Base(dirPath))
	}

	// create new directory object. (parent is not set)
	newDir := svc.NewDirectory(filepath.Base(dirPath), c.UserID, c.DriveID, dirPath)

	// see if the parent for this directory is already known.
	// if it is, then add this new directory as a subdirectory there,
	// otherwise place new directory under sfs root.
	parent := filepath.Dir(dirPath)
	pDir, err := c.Db.GetDirectoryByPath(parent)
	if err != nil {
		return err
	}
	if pDir != nil {
		newDir.Parent = pDir
	} else {
		newDir.Parent = c.Drive.Root
	}

	// add directory to service
	if err := c.Drive.AddSubDir(newDir.Parent.ID, newDir); err != nil {
		return err
	}
	if err := c.Db.AddDir(newDir); err != nil {
		return err
	}
	if err := c.WatchItem(dirPath); err != nil {
		return err
	}
	// push metadata to server if autosync is enabled
	if c.autoSync() {
		req, err := c.NewDirectoryRequest(newDir)
		if err != nil {
			return err
		}
		resp, err := c.Client.Do(req)
		if err != nil {
			return err
		}
		c.dump(resp, true)
	}
	return nil
}

// remove a directory from local and remote service instances.
func (c *Client) RemoveDir(dir *svc.Directory) error {
	if err := c.Drive.RemoveDir(dir.ID); err != nil {
		return err
	}
	if err := c.Db.RemoveDirectory(dir.ID); err != nil {
		return err
	}

	// TODO: remove files and subdirectories for this directory.
	//
	// need to think about this. this could easily be a recursive operation,
	// but there's a lot that needs to be accounted for if that's the route we want to go
	// subDirs := dir.GetSubDirs()
	// files := dir.GetFiles()

	return nil
}

func (c *Client) UpdateDirectory(updatedDir *svc.Directory) error {
	oldDir := c.Drive.GetDir(updatedDir.ID)
	if oldDir == nil {
		return fmt.Errorf("no such dir: %v", updatedDir.Name)
	}
	if err := c.Drive.UpdateDir(oldDir.ID, updatedDir); err != nil {
		return fmt.Errorf("failed to update directory in drive: %v", err)
	}
	if err := c.Db.UpdateDir(updatedDir); err != nil {
		return fmt.Errorf("failed to update directory in database: %v", err)
	}
	return nil
}

// get a directory using its SFS ID
func (c *Client) GetDirectoryByID(dirID string) (*svc.Directory, error) {
	dir := c.Drive.GetDir(dirID)
	if dir == nil {
		return nil, fmt.Errorf("directory %v not found", dirID)
	}
	return dir, nil
}

// get a directory object from the database using its path
func (c *Client) GetDirByPath(path string) (*svc.Directory, error) {
	dir, err := c.Db.GetDirectoryByPath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get directory: %v", err)
	} else if dir == nil {
		return nil, fmt.Errorf("directory does not exist: %s", path)
	}
	return dir, nil
}

// get a directory id from the DB using its file path
func (c *Client) GetDirIDFromPath(path string) (string, error) {
	dir, err := c.Db.GetDirectoryByPath(path)
	if err != nil {
		return "", fmt.Errorf("failed to get directory: %v", err)
	} else if dir == nil {
		return "", fmt.Errorf("directory does not exist: %s", path)
	}
	return dir.ID, nil
}

// ----- drive --------------------------------

// Loads drive from the database, populates root directory,
// and attaches to the client service instance.
func (c *Client) LoadDrive() error {
	drive, err := c.Db.GetDrive(c.DriveID)
	if err != nil {
		return err
	}
	if drive == nil {
		return fmt.Errorf("no drive found")
	}
	root, err := c.Db.GetDirectoryByID(drive.RootID)
	if err != nil {
		return err
	}
	if root == nil {
		return fmt.Errorf("no root directory associated with drive")
	}
	c.Drive = drive
	c.Drive.Root = c.Populate(root)
	c.Drive.IsLoaded = true
	if !c.Drive.IsIndexed() {
		c.Drive.SyncIndex = svc.BuildSyncIndex(c.Drive.Root)
	}
	return nil
}

// save drive metadata in the db
func (c *Client) SaveDrive(drv *svc.Drive) error {
	if err := c.Db.UpdateDrive(drv); err != nil {
		return fmt.Errorf("failed to update drive in database: %v", err)
	}
	if err := c.SaveState(); err != nil {
		return fmt.Errorf("failed to save state: %v", err)
	}
	return nil
}

// register a new drive with the server. if drive is already known to the server,
// then the server response should reflect this.
func (c *Client) RegisterClient() error {
	if c.Drive == nil {
		return fmt.Errorf("no drive available")
	}
	// register the user
	req, err := c.NewUserRequest(c.User)
	if err != nil {
		return fmt.Errorf("failed to create new user request: %v", err)
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		c.log.Warn(fmt.Sprintf("client failed to make request: %v", err))
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		c.log.Warn(fmt.Sprintf("failed to register new user. server status: %v", resp.Status))
		c.dump(resp, true)
		return nil
	}
	// register the drive. this will create a
	// serer-side root and allocate the server-side physical
	// drive directories and service files.
	req, err = c.NewDriveRequest(c.Drive)
	if err != nil {
		return fmt.Errorf("failed to create new drive request: %v", err)
	}
	resp, err = c.Client.Do(req)
	if err != nil {
		c.log.Warn(fmt.Sprintf("client failed to make request: %v", err))
		return nil
	}
	if resp.StatusCode == http.StatusOK {
		c.Drive.Registered = true
	} else {
		c.log.Warn(fmt.Sprintf("failed to register new drive. server status: %v", resp.Status))
		c.dump(resp, true)
		return nil
	}
	if err := c.SaveState(); err != nil {
		return fmt.Errorf("failed to save state: %v", err)
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
	root.Walk()

	// send everything to the database
	files := root.WalkFs()
	for _, file := range files {
		if err := c.Db.AddFile(file); err != nil {
			return nil, fmt.Errorf("failed to add file to database: %v", err)
		}
		if err := c.WatchItem(file.Path); err != nil {
			return nil, err
		}
	}
	dirs := root.WalkDs()
	for _, subDir := range dirs {
		if err := c.Db.AddDir(subDir); err != nil {
			return nil, fmt.Errorf("failed to add directory to database: %v", err)
		}
		if err := c.WatchItem(subDir.Path); err != nil {
			return nil, err
		}
	}
	// add root directory itself
	if err := c.Db.AddDir(root); err != nil {
		return nil, fmt.Errorf("failed to add root to database: %v", err)
	}
	if err := c.WatchItem(root.Path); err != nil {
		return nil, err
	}
	return root, nil
}

// similar to Discover, but uses a specified directory path
// and does not return a new directory object.
func (c *Client) DiscoverWithPath(dirPath string) error {
	// see if we have this directory already
	dir, err := c.Db.GetDirectoryByPath(dirPath)
	if err != nil {
		return err
	}
	if dir != nil {
		c.log.Info(fmt.Sprintf("directory %s is already known", filepath.Base(dirPath)))
		return nil
	}

	// create a new directory object under root and traverse
	log.Printf("[TEST] traversing %s...", dirPath)
	newDir := svc.NewDirectory(filepath.Base(dirPath), c.UserID, c.DriveID, dirPath)
	newDir.Parent = c.Drive.Root
	newDir.Walk()

	// add newly discovered files and directories to the service
	files := newDir.GetFiles()
	log.Printf("[TEST] adding %d files to the database...", len(files))

	for _, file := range files {
		if err := c.Db.AddFile(file); err != nil {
			return fmt.Errorf("failed to add file to database: %v", err)
		}
		if err := c.WatchItem(file.Path); err != nil {
			return err
		}
	}

	dirs := newDir.GetSubDirs()
	log.Printf("[TEST] adding %d directories...", len(dirs))

	for _, subDir := range dirs {
		if err := c.Db.AddDir(subDir); err != nil {
			return fmt.Errorf("failed to add directory to database: %v", err)
		}
		if err := c.WatchItem(subDir.Path); err != nil {
			return err
		}
	}

	// add new directory itself
	log.Printf("[TEST] adding %s...", filepath.Base(dirPath))
	if err := c.Db.AddDir(newDir); err != nil {
		return fmt.Errorf("failed to add root to database: %v", err)
	}
	if err := c.WatchItem(newDir.Path); err != nil {
		return err
	}
	return nil
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
		c.log.Warn("can't traverse directory without a path")
		return root
	}
	return c.populate(root)
}

func (c *Client) populate(dir *svc.Directory) *svc.Directory {
	entries, err := os.ReadDir(dir.Path)
	if err != nil {
		c.log.Error(err.Error())
		return dir
	}
	if len(entries) == 0 {
		return dir
	}
	for _, entry := range entries {
		entryPath := filepath.Join(dir.Path, entry.Name())
		item, err := os.Stat(entryPath)
		if err != nil {
			c.log.Error(fmt.Sprintf("could not get stat for entry %s - %v", entryPath, err))
			return dir
		}
		// add directory then recurse
		if item.IsDir() {
			subDir, err := c.Db.GetDirectoryByName(item.Name())
			if err != nil {
				c.log.Error(fmt.Sprintf("could not get directory (%s) from db:  %v", item.Name(), err))
				continue
			}
			if subDir == nil {
				continue
			}
			subDir = c.populate(subDir)
			if err := dir.AddSubDir(subDir); err != nil {
				c.log.Error(fmt.Sprintf("could not add directory: %v", err))
				continue
			}
		} else { // add file
			file, err := c.Db.GetFileByName(item.Name())
			if err != nil {
				c.log.Error(fmt.Sprintf("could not get file (%s) from db: %v", item.Name(), err))
				continue
			}
			if file == nil {
				continue
			}
			if err := dir.AddFile(file); err != nil {
				c.log.Error(fmt.Sprintf("could not add file (%s) to service: %v", item.Name(), err))
			}
		}
	}
	return dir
}

// recursively descends the drive's directory tree and compares what it
// finds to what is in the database, adding new items as it goes. generates
// a new root directory object and attaches it to the drive.
func (c *Client) RefreshDrive() {
	c.Drive.Root = c.refreshDrive(c.Drive.Root)
}

// descends users sfs directory tree and compares what it finds
// to what is in the database. if a new file or directory is found
// along the way it will be added to the database and new objects
// will created for them.
//
// does not account for files or directories not stored in the sfs file system!
func (c *Client) refreshDrive(dir *svc.Directory) *svc.Directory {
	entries, err := os.ReadDir(dir.Path)
	if err != nil {
		c.log.Error(fmt.Sprintf("refresh drive failed to read directory %s: %v", dir.Path, err))
		return dir
	}
	if len(entries) == 0 {
		return dir
	}
	for _, entry := range entries {
		entryPath := filepath.Join(dir.Path, entry.Name())
		item, err := os.Stat(entryPath)
		if err != nil {
			c.log.Error(fmt.Sprintf("could not get stat for entry %s - %v", entryPath, err))
			return dir
		}
		// add directory then recurse
		if item.IsDir() {
			subDir, err := c.Db.GetDirectoryByPath(entryPath)
			if err != nil {
				c.log.Error(fmt.Sprintf("could not get directory (%s) from db: %v", item.Name(), err))
				continue
			}
			// new directory
			if subDir == nil {
				subDir = svc.NewDirectory(item.Name(), dir.OwnerID, dir.DriveID, entryPath)
				if err := c.Db.AddDir(subDir); err != nil {
					c.log.Error(fmt.Sprintf("could not add directory (%s) to db: %v", item.Name(), err))
					continue
				}
				if err := c.WatchItem(entryPath); err != nil {
					c.log.Error(fmt.Sprintf("could not monitor directory (%s): %v", item.Name(), err))
					continue
				}
				subDir = c.refreshDrive(subDir)
				if err := dir.AddSubDir(subDir); err != nil {
					log.Print(err)
				}
			}
		} else {
			file, err := c.Db.GetFileByPath(entryPath)
			if err != nil {
				c.log.Error(fmt.Sprintf("could not get file (%s) from db: %v", item.Name(), err))
				continue
			}
			// new file
			if file == nil {
				newFile := svc.NewFile(item.Name(), dir.DriveID, dir.OwnerID, entryPath)
				newFile.DirID = dir.ID
				if err := c.Db.AddFile(newFile); err != nil {
					c.log.Error(fmt.Sprintf("could not add file (%s) to db: %v", item.Name(), err))
					continue // TEMP until there's a better way to handle this error
				}
				if err := dir.AddFile(newFile); err != nil {
					c.log.Error(fmt.Sprintf("could not add file (%s) service: %v", item.Name(), err))
				}
				if err := c.WatchItem(entryPath); err != nil {
					c.log.Error(fmt.Sprintf("failed to watch file (%s): %v", item.Name(), err))
					continue
				}
			}
		}
	}
	return dir
}
