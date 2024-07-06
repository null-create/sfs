package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sfs/pkg/logger"
	"github.com/sfs/pkg/monitor"
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
		// NOTE: directory actions aren't supported yet
		// dir, err := c.Db.GetDirectoryByPath(itemPath)
		// if err != nil {
		// 	return err
		// }
		// if err := c.RemoveDir(dir); err != nil {
		// 	return err
		// }
	} else {
		file, err := c.Db.GetFileByPath(itemPath)
		if err != nil {
			return nil
		}
		if file == nil {
			return fmt.Errorf("%s not found", filepath.Base(itemPath))
		}
		if err := c.RemoveFile(file); err != nil {
			return err
		}
	}
	return nil
}

// send item metadata to the server. used with
// event handler loop.
func (c *Client) UpdateItem(item monitor.EItem) error {
	if item.IsDir() {
		dir, err := c.GetDirByPath(item.Path())
		if err != nil {
			return err
		}
		req, err := c.UpdateDirectoryRequest(dir)
		if err != nil {
			return err
		}
		resp, err := c.Client.Do(req)
		if err != nil {
			c.log.Error(fmt.Sprintf("failed to execute request: %v", err))
			return err
		}
		c.dump(resp, true)
	} else {
		file, err := c.GetFileByPath(item.Path())
		if err != nil {
			return err
		}
		req, err := c.UpdateFileRequest(file)
		if err != nil {
			return err
		}
		resp, err := c.Client.Do(req)
		if err != nil {
			c.log.Error(fmt.Sprintf("failed to execute request: %v", err))
			return err
		}
		c.dump(resp, true)
	}
	return nil
}

// is this file or directory already registered with client *and* the server?
func (c *Client) KnownItem(itemPath string) bool {
	item, err := os.Stat(itemPath)
	if err != nil {
		log.Fatal(err)
	}
	if item.IsDir() {
		// make sure this is registered with the client first
		d, err := c.GetDirByPath(itemPath)
		if err != nil {
			c.log.Error(err.Error())
			return false
		}
		if d == nil {
			c.log.Error(fmt.Sprintf("%s is not registered with the client", filepath.Base(itemPath)))
			return false
		}
		return true
	} else {
		// make sure this is registered with the client first
		f, err := c.GetFileByPath(itemPath)
		if err != nil {
			c.log.Error(err.Error())
			return false
		}
		if f == nil {
			c.log.Error(fmt.Sprintf("%s is not registered with the client", filepath.Base(itemPath)))
			return false
		}
		return true
	}
}

// send file or directory metadata to the server to register.
func (c *Client) RegisterItem(itemPath string) error {
	if !c.KnownItem(itemPath) {
		item, err := os.Stat(itemPath)
		if err != nil {
			return err
		}
		if item.IsDir() {
			d, err := c.GetDirByPath(itemPath)
			if err != nil {
				return err
			}
			if d == nil {
				c.log.Error(fmt.Sprintf("%s is not registered with client", filepath.Base(itemPath)))
			}
			if err := c.RegisterDirectory(d); err != nil {
				return err
			}
		} else {
			f, err := c.GetFileByPath(itemPath)
			if err != nil {
				return err
			}
			if f == nil {
				c.log.Error(fmt.Sprintf("%s is not registered with client", filepath.Base(itemPath)))
				return nil
			}
			if err := c.RegisterFile(f); err != nil {
				return err
			}
		}
	}
	return nil
}

// find a file or directory by name.
//
// NOTE: this could cause collisions because files are
// stored in the db with unique *ids,* not names
func (c *Client) GetItemByName(itemName string) *Item {
	thing, err := os.Stat(itemName)
	if err != nil {
		log.Fatal(err)
	}
	var item = new(Item)
	if thing.IsDir() {
		dir, err := c.GetDirByName(itemName)
		if err != nil {
			c.log.Error(err.Error())
		}
		if dir == nil {
			c.log.Error(fmt.Sprintf("%s not found", itemName))
		}
		item.Directory = dir
	} else {
		file, err := c.GetFileByName(itemName)
		if err != nil {
			c.log.Error(err.Error())
		}
		if file == nil {
			c.log.Error(fmt.Sprintf("%s not found found", itemName))
		}
		item.File = file
	}
	return item
}

// finds an item by path. returns nil if not found.
func (c *Client) GetItemByPath(path string) (*Item, error) {
	thing, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}
	var item = new(Item)
	if thing.IsDir() {
		dir, err := c.GetDirByPath(path)
		if err != nil {
			c.log.Error(err.Error())
			return nil, err
		}
		if dir == nil {
			c.log.Info(fmt.Sprintf("%s not found", filepath.Base(path)))
			return nil, nil
		}
		item.Directory = dir
	} else {
		file, err := c.GetFileByPath(path)
		if err != nil {
			c.log.Error(err.Error())
			return nil, err
		}
		if file == nil {
			c.log.Info(fmt.Sprintf("%s not found found", filepath.Base(path)))
			return nil, nil
		}
		item.File = file
	}
	return item, nil
}

// ----- files --------------------------------------

// does this file or directory exist?
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
	files := c.Drive.GetFilesMap()
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

/*
add a file to the service using its file path.
should check for whether the directory it resides in is
monitored by SFS -- though is not contingent on it!

SFS can monitor files outside of the designated root directory, so
if we add a file this way then we should automatically make a backup of it
in the SFS server.
*/
func (c *Client) AddFile(filePath string) error {
	// see if we already have this file in the system
	file, err := c.Db.GetFileByPath(filePath)
	if err != nil {
		return err
	}
	if file != nil {
		return fmt.Errorf("%s is already registered", filepath.Base(filePath))
	}

	// create new file object
	newFile := svc.NewFile(filepath.Base(filePath), c.DriveID, c.UserID, filePath)

	// see if we already have the file's parent directory in the file system
	dir, err := c.GetDirByPath(filepath.Dir(filePath))
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
	// this will create an intial EMPTY file on the server-side.
	// backup contents are created during the first sync of the file
	// after being registered.
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
		// get newly generated server-side path for the file if successfully created
		if resp.StatusCode == http.StatusOK {
			svrpath, err := c.getFileServerPath(newFile)
			if err != nil {
				return err
			}
			newFile.ServerPath = svrpath
			if err := c.UpdateFile(newFile); err != nil {
				return err
			}
		} else {
			c.log.Warn(fmt.Sprintf("failed to send metadata to server: %v", resp.Status))
		}
	}
	c.log.Info(fmt.Sprintf("added %s to client", newFile.Name))
	return nil
}

// retrieve the updated server path for the file after a successful
// registration with the server.
func (c *Client) getFileServerPath(file *svc.File) (string, error) {
	req, err := c.GetFileInfoRequest(file)
	if err != nil {
		return "", err
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get server path. server returned non-200 status: %v", resp.Status)
	}
	// get file info from response so we can parse it for the new server path
	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return "", err
	}
	resp.Body.Close()

	var f = new(svc.File)
	if err := json.Unmarshal(buf.Bytes(), &f); err != nil {
		return "", err
	}
	if f.ServerPath == "" || f.ServerPath == f.ClientPath {
		return "", fmt.Errorf("server path was not set correctly. client_path=%s server_path=%v", f.ClientPath, f.ServerPath)
	}
	return f.ServerPath, nil
}

// add a new file to a specified directory using a directory ID.
// adds file to database and monitoring services.
func (c *Client) AddFileWithDirID(dirID string, newFile *svc.File) error {
	if err := c.Drive.AddFile(dirID, newFile); err != nil {
		return err
	}
	if err := c.Db.AddFile(newFile); err != nil {
		return err
	}
	if err := c.WatchItem(newFile.ClientPath); err != nil {
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
		// get newly generated server-side path for the file if successfully created
		if resp.StatusCode == http.StatusOK {
			svrpath, err := c.getFileServerPath(newFile)
			if err != nil {
				return err
			}
			newFile.ServerPath = svrpath
			if err := c.UpdateFile(newFile); err != nil {
				return err
			}
		}
		c.dump(resp, true)
	}
	c.log.Info(fmt.Sprintf("added %s to client", newFile.Name))
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
	c.log.Info(fmt.Sprintf("%s was modified", file.Name))
	return nil
}

// update file metadata in the service instance. does not
// update file contents. use c.ModifyFile() instead.
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

// remove a file in a specied directory. removes the file from the server too.
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
	c.log.Info(fmt.Sprintf("%s was moved to the recycle bin", file.Name))

	// remove file from the server if auto sync is enabled.
	if c.autoSync() {
		req, err := c.DeleteFileRequest(file)
		if err != nil {
			c.log.Error("failed to create request: " + err.Error())
			return nil
		}
		resp, err := c.Client.Do(req)
		if err != nil {
			c.log.Error("failed to execute HTTP request: " + err.Error())
			return nil
		}
		c.dump(resp, true)
	}
	return nil
}

// move a file from one location to another on a users computer via
// the sfs client. set keepOrig to true to keep the original copy in the original location.
// sfs will only monitor the new copy after the move.
//
// destPath must be the absolute path for the files destination (i.e., end with the file name)
func (c *Client) MoveFile(destPath string, filePath string, keepOrig bool) error {
	file, err := c.GetFileByPath(filePath)
	if err != nil {
		return err
	}
	if file == nil {
		return fmt.Errorf("file not found: " + filepath.Base(filePath))
	}
	if filePath != file.ClientPath {
		return fmt.Errorf("source path does not match original file path. fp=%q orig=%s", filePath, file.ClientPath)
	}

	// copy physical file
	var origPath = file.ClientPath
	if err := file.Copy(destPath); err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}
	file.ClientPath = destPath

	// update dbs
	if err := c.Db.UpdateFile(file); err != nil {
		return fmt.Errorf("failed to update file database: %v", err)
	}
	c.log.Info(fmt.Sprintf("%s moved from %s to %s", file.Name, origPath, destPath))
	return nil
}

// see if this file is registered with the server (exists on servers DB)
func (c *Client) IsFileRegistered(file *svc.File) (bool, error) {
	req, err := c.GetFileInfoRequest(file)
	if err != nil {
		return false, fmt.Errorf("failed to create file request: %v", err)
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to execute request: %v", err)
	}
	fmt.Printf("response status: %v", resp.Status)
	return resp.StatusCode == http.StatusOK, nil
}

// register new file with the server. does not send file contents,
// only metadata
func (c *Client) RegisterFile(file *svc.File) error {
	req, err := c.NewFileRequest(file)
	if err != nil {
		return err
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	c.dump(resp, true)
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
	c.log.Info(fmt.Sprintf("directory (%s) added to client", dir.Name))
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
	// NOTE: directory monitoring is not currently supported.
	// keeping this for future implementation iterations.
	// if err := c.WatchItem(dirPath); err != nil {
	// 	return err
	// }
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
		// TODO: get directory server path.

		c.dump(resp, true)
	}
	c.log.Info(fmt.Sprintf("directory (%s) added to client", newDir.Name))
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
	c.log.Info(fmt.Sprintf("directory (%s) updated", updatedDir.Name))
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

// get a directory id from the DB using its file path.
// returns an error if the directory is not found.
func (c *Client) GetDirIDFromPath(path string) (string, error) {
	dir, err := c.Db.GetDirectoryByPath(path)
	if err != nil {
		return "", fmt.Errorf("failed to get directory: %v", err)
	} else if dir == nil {
		return "", fmt.Errorf("directory does not exist: %s", path)
	}
	return dir.ID, nil
}

// search for a directory by name.
// returns an error if the directory does not exist
func (c *Client) GetDirByName(name string) (*svc.Directory, error) {
	if name == "" {
		return nil, fmt.Errorf("no path specified")
	}
	dir, err := c.Db.GetDirectoryByName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get directory: %v", err)
	}
	if dir == nil {
		return nil, fmt.Errorf(fmt.Sprintf("directory does not exist: %s", name))
	}
	return dir, nil
}

// see if this directory is registered with the server (exists on servers DB)
func (c *Client) IsDirRegistered(dir *svc.Directory) bool {
	req, err := c.GetDirInfoRequest(dir)
	if err != nil {
		c.log.Error("failed to create directory request: " + err.Error())
		return false
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		c.log.Error("failed to execute request: " + err.Error())
		return false
	}
	if resp.StatusCode == http.StatusOK {
		return true
	}
	return false
}

// send directory metadata to the server
func (c *Client) RegisterDirectory(dir *svc.Directory) error {
	req, err := c.GetDirReq(dir, "NEW")
	if err != nil {
		return err
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		c.log.Warn(fmt.Sprintf("server dir registration response: %d", resp.StatusCode))
	} else {
		c.log.Log("INFO", fmt.Sprintf("directory %s registered with server", dir.Name))
	}
	return nil
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

	// populate with the contents of the designated root directory
	c.Drive.Root = c.Populate(root)

	// add all other distributed files and subdirectories monitored by sfs
	files, err := c.Db.GetUsersFiles(c.UserID)
	if err != nil {
		return err
	}
	c.Drive.Root.AddFiles(files)

	dirs, err := c.Db.GetUsersDirectories(c.UserID)
	if err != nil {
		return err
	}
	if err := c.Drive.Root.AddSubDirs(dirs); err != nil {
		return err
	}
	c.Drive.IsLoaded = true

	// build client sync index
	c.BuildSyncIndex()

	c.log.Log("INFO", "drive loaded")
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
	c.log.Info("drive saved")
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
	c.log.Info("client registered with the server")
	if err := c.SaveState(); err != nil {
		return fmt.Errorf("failed to save state: %v", err)
	}
	return nil
}

/*
Iterate over ALL users files in the client side DBs and see if
there are any that aren't registered with the server.

If there's some that aren't, prompt the user whether they want
to push them to the server. If yes, push non-registered files
to the server.
*/
func (c *Client) Refresh() error {
	files, err := c.Db.GetUsersFiles(c.UserID)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		c.log.Log(logger.INFO, "no files registered with client. nothing to refresh")
		return nil
	}

	// see if any of these aren't registered with the server
	var toRegister = make([]*svc.File, 0, len(files))
	for _, file := range files {
		reg, err := c.IsFileRegistered(file)
		if err != nil {
			c.log.Error(err.Error())
		}
		if !reg {
			toRegister = append(toRegister, file)
		}
	}

	c.log.Info(fmt.Sprintf("%d files need to be registered with the server", len(toRegister)))
	if c.Continue() {
		c.log.Log(logger.INFO, fmt.Sprintf("registering %d files with the server...", len(toRegister)))
		for _, file := range toRegister {
			if err := c.RegisterFile(file); err != nil {
				c.log.Error("failed to register file: " + err.Error())
			}
		}
	}
	return nil
}

/*
TODO:

Query server for all the users file and compare with the local
client-side DB. If there are any on the server that aren't on the client
side, pull those files from the server and add to the local client service.
*/
func (c *Client) RefreshFromServer() error { return nil }

// similar to DiscoverInRoot, but uses a specified directory path
// and does not return a new directory object.
func (c *Client) DiscoverWithPath(dirPath string) (*svc.Directory, error) {
	// see if we have this directory already
	dir, err := c.Db.GetDirectoryByPath(dirPath)
	if err != nil {
		return nil, err
	}
	if dir != nil {
		c.log.Info(fmt.Sprintf("%s is already known", filepath.Base(dirPath)))
		return dir, nil
	}

	// create a new directory object and traverse
	c.log.Info(fmt.Sprintf("traversing %s...", dirPath))
	newDir := svc.NewDirectory(filepath.Base(dirPath), c.UserID, c.DriveID, dirPath)
	newDir.Parent = c.Drive.Root
	newDir.Walk()

	// add newly discovered files and directories to the service
	files := newDir.GetFiles()
	c.log.Info(fmt.Sprintf("adding %d files...", len(files)))

	if err := c.Db.AddFiles(files); err != nil {
		return nil, fmt.Errorf("failed to add files to database: %v", err)
	}
	for _, file := range files {
		if err := c.WatchItem(file.Path); err != nil {
			return nil, err
		}
		if c.autoSync() {
			if err := c.RegisterFile(file); err != nil {
				return nil, err
			}
		}
	}

	// add directories to the database. not monitored (for now)
	dirs := newDir.GetSubDirs()
	c.log.Info(fmt.Sprintf("adding %d directories...", len(dirs)))

	if err := c.Db.AddDirs(dirs); err != nil {
		return nil, err
	}
	for _, subDir := range dirs {
		// if err := c.WatchItem(subDir.Path); err != nil {
		// 	return err
		// }
		if c.autoSync() {
			if err := c.RegisterDirectory(subDir); err != nil {
				return nil, err
			}
		}
	}

	// add new directory itself. not monitored (for now)
	c.log.Info(fmt.Sprintf("adding %s...", filepath.Base(dirPath)))
	if err := c.Db.AddDir(newDir); err != nil {
		return nil, fmt.Errorf("failed to add root to database: %v", err)
	}
	// if err := c.WatchItem(newDir.Path); err != nil {
	// 	return err
	// }
	if err := c.RegisterDirectory(newDir); err != nil {
		return nil, err
	}
	return newDir, nil
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
					c.log.Error(fmt.Sprintf("failed to add directory (%s): %v", subDir.ID, err))
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
