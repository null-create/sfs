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
	"strconv"
	"strings"

	"github.com/sfs/pkg/logger"
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

// remove an item from the local SFS service
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
		if dir == nil {
			return fmt.Errorf("dir '%s' not found", filepath.Base(itemPath))
		}
		if err := c.RemoveDir(dir); err != nil {
			return err
		}
	} else {
		file, err := c.Db.GetFileByPath(itemPath)
		if err != nil {
			return nil
		}
		if file == nil {
			return fmt.Errorf("file '%s' not found", filepath.Base(itemPath))
		}
		if err := c.RemoveFile(file); err != nil {
			return err
		}
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
			c.log.Info(fmt.Sprintf("'%s' not found", filepath.Base(path)))
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
			c.log.Info(fmt.Sprintf("'%s' not found found", filepath.Base(path)))
			return nil, nil
		}
		item.File = file
	}
	return item, nil
}

// ------ configuration --------------------------------

// update user-specific settings
func (c *Client) UpdateConfigSetting(setting, value string) error {
	switch setting {
	case "CLIENT_NAME":
		return c.updateClientName(value)
	case "CLIENT_EMAIL":
		return c.updateClientEmail(value)
	case "CLIENT_PASSWORD":
		return c.updateUserPassword(value, c.User.Password)
	case "CLIENT_PORT":
		return c.updateClientPort(value)
	case "CLIENT_BACKUP_DIR":
		return c.UpdateBackupPath(value)
	case "CLIENT_LOCAL_BACKUP":
		return c.SetLocalBackup(value)
	case "CLIENT_NEW_SERVICE":
		return envCfgs.Set(setting, value)
	case "NEW_SERVICE":
		return envCfgs.Set(setting, value)
	default:
		return fmt.Errorf("unknown setting: '%s'", setting)
	}
}

// enable or disable backing up files to local storage.
func (c *Client) SetLocalBackup(modeStr string) error {
	mode, err := strconv.ParseBool(modeStr)
	if err != nil {
		c.log.Error(fmt.Sprintf("failed to parse string: %v", err))
		return err
	}
	c.Conf.LocalBackup = mode
	if err := envCfgs.Set("CLIENT_LOCAL_BACKUP", strconv.FormatBool(mode)); err != nil {
		return err
	}
	if err := c.SaveState(); err != nil {
		c.log.Error("failed to update state file: " + err.Error())
	} else {
		if mode {
			c.log.Info("local backup enabled")
		} else {
			c.log.Info("local backup disabled")
		}
	}
	return nil
}

// update the backup configurations for the service and all client-side
// files and directories
func (c *Client) UpdateBackupPath(newDirPath string) error {
	if !c.isDirPath(newDirPath) {
		c.log.Error("path is not a directory")
		return fmt.Errorf("path is not a directory")
	}
	if c.LocalBackupDir == newDirPath {
		return nil // nothing to replace
	}
	return c.updateBackupPaths(filepath.Clean(newDirPath))
}

// If a user specifies a custom backup directory, then all
// local backup items need to have their backup paths updated.
// we don't want files to be backed up to a default directory if
// a user tries to specify otherwise.
func (c *Client) updateBackupPaths(newPath string) error {
	oldPath := c.Root
	dirs := c.Drive.GetDirs()
	for _, dir := range dirs {
		dir.BackupPath = strings.Replace(dir.BackupPath, oldPath, newPath, 1)
	}
	if err := c.Db.UpdateDirs(dirs); err != nil {
		return err
	}
	files := c.Drive.GetFiles()
	for _, file := range files {
		file.BackupPath = strings.Replace(file.BackupPath, oldPath, newPath, 1)
	}
	if err := c.Db.UpdateFiles(files); err != nil {
		return err
	}
	c.Conf.BackupDir = newPath
	if err := envCfgs.Set("CLIENT_BACKUP_DIR", newPath); err != nil {
		return err
	}
	if err := c.SaveState(); err != nil {
		return err
	}
	return nil
}

// update user's name
func (c *Client) updateClientName(newName string) error {
	if newName == c.Conf.User && newName == c.User.Name && newName == c.Drive.OwnerName {
		return nil
	}
	user, err := c.Db.GetUser(c.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found: %v", c.UserID)
	}
	c.User.Name = newName
	c.Conf.User = newName
	c.Drive.OwnerName = newName
	user.Name = newName
	if err := c.Db.UpdateUser(user); err != nil {
		return err
	}
	if err := c.Db.UpdateDrive(c.Drive); err != nil {
		return err
	}
	if err := envCfgs.Set("CLIENT_NAME", newName); err != nil {
		return err
	}
	if err := c.SaveState(); err != nil {
		return err
	}
	return nil
}

// TODO: update all items owner name in the DB with the new user's new name
func (c *Client) updateFileOwnerName(newName string) error {
	return nil
}

// update user's email
func (c *Client) updateClientEmail(newEmail string) error {
	user, err := c.Db.GetUser(c.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found: '%s' (id=%v)", c.Conf.User, c.UserID)
	}
	c.User.Email = newEmail
	user.Email = newEmail
	if err := c.Db.UpdateUser(user); err != nil {
		return err
	}
	if err := envCfgs.Set("CLIENT_EMAIL", newEmail); err != nil {
		return err
	}
	if err := c.SaveState(); err != nil {
		return err
	}
	return nil
}

// update user's password
func (c *Client) updateUserPassword(oldPw, newPw string) error {
	user, err := c.Db.GetUser(c.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found: '%s' (id=%v)", c.Conf.User, c.UserID)
	}
	if newPw == user.Password && newPw == c.User.Password {
		return nil // nothing to update
	}
	// make sure current password is valid
	if oldPw != user.Password && oldPw != c.User.Password {
		return fmt.Errorf("incorrect password. password not updated")
	}
	user.Password = newPw
	c.User.Password = newPw
	// TODO: hashing of user passwords should occur before saving to DB.
	// dont save them as plaintext!
	if err := c.Db.UpdateUser(user); err != nil {
		return err
	}
	if err := envCfgs.Set("CLIENT_PASSWORD", newPw); err != nil {
		return err
	}
	if err := c.SaveState(); err != nil {
		return err
	}
	return nil
}

// update client port setting
func (c *Client) updateClientPort(pvalue string) error {
	port, err := strconv.Atoi(pvalue)
	if err != nil {
		return err
	}
	if c.Conf.Port == port {
		return nil // nothing to do here
	}
	c.Conf.Port = port
	return envCfgs.Set("CLIENT_PORT", pvalue)
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
	var output string
	files := c.Drive.GetFilesMap()
	for _, f := range files {
		output += fmt.Sprintf("id: %s\nname: %s\nloc: %s\n\n", f.ID, f.Name, f.ClientPath)
	}
	fmt.Print(output)
}

// list all files managed by the local sfs database
func (c *Client) ListLocalFilesDB() error {
	files, err := c.Db.GetUsersFiles(c.UserID)
	if err != nil {
		return err
	}
	var output string
	for _, f := range files {
		output += fmt.Sprintf("id: %s\nname: %s\nloc: %s\n\n", f.ID, f.Name, f.ClientPath)
	}
	fmt.Print(output)
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
		return nil, fmt.Errorf("failed to get file by path from db: %v", err)
	} else if file == nil {
		return nil, fmt.Errorf("'%s' not found", filepath.Base(path))
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

SFS can monitor files outside of the designated root directory, so
if we add a file this way then we should automatically make a backup of it
in the SFS server, or locally to a designated directory, depending
on the service configurations.
*/
func (c *Client) AddFile(filePath string) error {
	// see if we already have this file in the system
	file, err := c.Db.GetFileByPath(filePath)
	if err != nil {
		return err
	}
	if file != nil {
		return fmt.Errorf("'%s' is already registered", filepath.Base(filePath))
	}

	// create new file object
	newFile := svc.NewFile(filepath.Base(filePath), c.Drive.ID, c.UserID, filePath)

	// see if we already have the file's parent directory in the file system
	dir, err := c.GetDirByPath(filepath.Dir(filePath))
	if err != nil && strings.Contains(err.Error(), "does not exist") {
		// if the parent directory to this file doesn't exist in the file system,
		// then add it to the SFS root.
		newFile.DirID = c.Drive.Root.ID
	} else if err != nil {
		return err
	} else {
		newFile.DirID = dir.ID
	}

	// mark this file as being backed up by the service
	newFile.MarkLocalBackup()

	// add file to sfs system
	// NOTE: backup paths are generated when adding the new file to the directory.
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
	if !c.localBackup() {
		// push metadata to server if autosync is enabled
		// and we don't default to using local storage for backup purposes.
		//
		// this will create an intial EMPTY file on the server-side.
		// backup contents are created during the first sync of the file
		// after being registered.
		req, err := c.NewFileRequest(newFile)
		if err != nil {
			return err
		}
		resp, err := c.Client.Do(req)
		if err != nil {
			return err
		}
		c.dump(resp, true)
		if resp.StatusCode == http.StatusOK {
			// update client side info about the file to
			// include the server path generated after a successful registration
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
	} else {
		// make a local backup copy of the file
		if err := c.BackupFile(newFile); err != nil {
			return err
		}
	}
	// update service state
	if err := c.SaveState(); err != nil {
		c.log.Error(fmt.Sprintf("failed to save state file: %v", err))
	}
	c.log.Info(fmt.Sprintf("file '%s' added to client", newFile.Name))
	return nil
}

// retrieve the updated server path for the file after a successful
// registration with the server.//
// returns an empty string if the client failed to make contact with the server.
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
		c.log.Warn(fmt.Sprintf("failed to get server path. server returned non-200 status: %v", resp))
		return "", fmt.Errorf(fmt.Sprintf("failed to get server path. server returned non-200 status: %v", resp))
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
	if !c.localBackup() {
		// push metadata to server
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
	} else {
		if err := c.BackupFile(newFile); err != nil {
			return err
		}
	}
	c.log.Info(fmt.Sprintf("added %s to client", newFile.Name))
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

// remove a file.
// removes the file from the server if local backup is disabled.
func (c *Client) RemoveFile(file *svc.File) error {
	// make sure this file is actually registered with the service
	// before mucking around.
	if !c.KnownItem(file.ClientPath) {
		return fmt.Errorf("file '%s' not registered", file.Name)
	}
	// stop monitoring the file
	c.Monitor.StopWatching(file.ClientPath)
	// move the file to the SFS recycle bin to help with recovery in case
	// of an accidental deletion.
	if err := file.Copy(filepath.Join(c.RecycleBin, file.Name)); err != nil {
		return fmt.Errorf("failed to copy file to recyle directory: %v", err)
	}
	// remove file data from the service. does not remove physical file,
	// only data used by the service.
	if err := c.Drive.RemoveFile(file.DirID, file); err != nil {
		return err
	}
	if err := c.Db.RemoveFile(file.ID); err != nil {
		return err
	}
	c.log.Info(fmt.Sprintf("%s was moved to the recycle bin", file.Name))
	// remove file from the server if local backup is disabled
	if !c.localBackup() {
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
	if resp.StatusCode == http.StatusInternalServerError {
		c.dump(resp, true)
	} else if resp.StatusCode == http.StatusOK {
		c.log.Info(fmt.Sprintf("file '%s' registered", file.Name))
	}
	return nil
}

// ----- directories --------------------------------

func (c *Client) IsDir(path string) bool {
	item, err := os.Stat(path)
	if err != nil {
		c.log.Error(fmt.Sprintf("failed to get stat for item: %v\n%v", path, err))
	}
	return item.IsDir()
}

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
	var output string
	for _, dir := range dirs {
		output += fmt.Sprintf(
			"name: %s\nid: %v\nloc: %s\n\n",
			dir.Name, dir.ID, dir.ClientPath,
		)
	}
	fmt.Print(output)
	return nil
}

// get the directories server path.
func (c *Client) getDirServerPath(dir *svc.Directory) (string, error) {
	dirReq, err := c.GetDirInfoRequest(dir)
	if err != nil {
		return "", err
	}
	res, err := c.Client.Do(dirReq)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received a non 200 response from server: %v", res)
	}
	var buf bytes.Buffer
	_, err = io.Copy(&buf, res.Body)
	if err != nil {
		return "", err
	}
	res.Body.Close()

	var d = new(svc.Directory)
	if err := json.Unmarshal(buf.Bytes(), &d); err != nil {
		return "", err
	}
	if d.ServerPath == "" || d.ServerPath == d.ClientPath {
		return "", fmt.Errorf("server path was not set correctly. client_path=%s server_path=%v", d.ClientPath, d.ServerPath)
	}
	return d.ServerPath, nil
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

	// see if the parent directory for this directory is already known.
	// if it is, then add this new directory as a subdirectory there,
	// otherwise place new directory under sfs root.
	pd := filepath.Dir(dirPath)
	parent, err := c.Db.GetDirectoryByPath(pd)
	if err != nil {
		return err
	}
	// add directory to service
	// NOTE: directory backup paths are updated when adding to the drive
	if parent != nil {
		if err := parent.AddSubDir(newDir); err != nil {
			return err
		}
	} else {
		if err := c.Drive.Root.AddSubDir(newDir); err != nil {
			return err
		}
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
	if !c.localBackup() {
		req, err := c.NewDirectoryRequest(newDir)
		if err != nil {
			return err
		}
		resp, err := c.Client.Do(req)
		if err != nil {
			return err
		}
		c.dump(resp, true)
		if resp.StatusCode == http.StatusOK {
			serverPath, err := c.getDirServerPath(newDir)
			if err != nil {
				return err
			}
			newDir.ServerPath = serverPath
			if err := c.UpdateDirectory(newDir); err != nil {
				return err
			}
		}
	} else {
		if err := c.BackupDir(newDir); err != nil {
			return err
		}
	}
	c.log.Info(fmt.Sprintf("directory (%s) added to client", newDir.Name))
	return nil
}

// remove a directory from local and remote service instances.
func (c *Client) RemoveDir(dirToRemove *svc.Directory) error {
	d, err := c.Db.GetDirectoryByPath(dirToRemove.Path)
	if err != nil {
		return err
	}
	if d == nil {
		return fmt.Errorf("directory '%s' not found", dirToRemove.ID)
	}
	// first remove all the files and subdirectories in this directory
	files := dirToRemove.GetFiles()
	if err := c.Db.RemoveFiles(files); err != nil {
		return err
	}
	subDirs := dirToRemove.GetSubDirs()
	if err := c.Db.RemoveDirectories(subDirs); err != nil {
		return err
	}
	// remove directory itself from the service
	if err := c.Drive.RemoveDir(dirToRemove.ID); err != nil {
		return err
	}
	if err := c.Db.RemoveDirectory(dirToRemove.ID); err != nil {
		return err
	}
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
	c.log.Info(fmt.Sprintf("directory '%s' (id=%s) updated", updatedDir.Name, updatedDir.ID))
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

// see if this directory is registered with the server (exists on servers DB)
func (c *Client) IsDirRegistered(dir *svc.Directory) (bool, error) {
	req, err := c.GetDirInfoRequest(dir)
	if err != nil {
		return false, fmt.Errorf("failed to create directory request: " + err.Error())
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to execute request: " + err.Error())
	}
	return resp.StatusCode == http.StatusOK, nil
}

// send directory metadata to the server
func (c *Client) RegisterDirectory(dir *svc.Directory) error {
	req, err := c.NewDirectoryRequest(dir)
	if err != nil {
		return err
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		c.log.Info(fmt.Sprintf("directory '%s' registered", dir.Name))
	} else {
		c.dump(resp, true)
	}
	return nil
}

func (c *Client) EmptyRecycleBin() error {
	c.log.Info("emptying client recycle bin...")
	entries, err := os.ReadDir(c.RecycleBin)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(c.RecycleBin, entry.Name())); err != nil {
			c.log.Error(err.Error())
		}
	}
	c.log.Info(fmt.Sprintf("client recycle bin emptied. %d files deleted", len(entries)))
	return nil
}

// ----- drive --------------------------------

// add a new drive to the client. mainly used for testing
func (c *Client) AddDrive(drv *svc.Drive) error {
	files := drv.Root.GetFiles()
	if err := c.Db.AddFiles(files); err != nil {
		return err
	}
	subDirs := drv.Root.GetSubDirs()
	if err := c.Db.AddDirs(subDirs); err != nil {
		return err
	}
	// add the root directory and the drive itself
	if err := c.Db.AddDir(drv.Root); err != nil {
		return err
	}
	if err := c.Db.AddDrive(drv); err != nil {
		return err
	}
	// finally, add the drive to the client instance and save client state
	c.Drive = drv
	if err := c.SaveState(); err != nil {
		return fmt.Errorf("failed to save state file: %v", err)
	}
	return nil
}

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
	drive.Root = root
	c.Drive = drive
	c.Drive.IsLoaded = true

	// add users monitored directories
	dirs, err := c.Db.GetUsersDirectories(c.UserID)
	if err != nil {
		return err
	}
	if err := c.Drive.Root.AddSubDirs(dirs); err != nil {
		return err
	}

	// add all other distributed files and subdirectories monitored by sfs
	files, err := c.Db.GetUsersFiles(c.UserID)
	if err != nil {
		return err
	}
	c.Drive.Root.AddFiles(files)

	// build client sync index
	c.BuildSyncIndex()

	c.log.Log(logger.INFO, "drive loaded")

	// update state file
	if err := c.SaveState(); err != nil {
		return err
	}
	return nil
}

// save drive metadata in the db
func (c *Client) SaveDrive(drv *svc.Drive) error {
	if c.DriveID == drv.ID {
		if err := c.Db.UpdateDrive(drv); err != nil {
			return fmt.Errorf("failed to update drive in database: %v", err)
		}
		c.Drive = drv
		if err := c.SaveState(); err != nil {
			return fmt.Errorf("failed to save state: %v", err)
		}
		c.log.Info(fmt.Sprintf("drive (id=%s) updated", drv.ID))
		return nil
	} else {
		return fmt.Errorf("drive mismatch. client drive id: %s given drive id: %s", c.DriveID, drv.ID)
	}
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
	} else if resp.StatusCode == http.StatusBadRequest {
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
	var toRegister = make([]*svc.File, 0)
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

// Does this path point to a directory?
func (c *Client) isDirPath(path string) bool {
	item, err := os.Stat(path)
	if err != nil {
		c.log.Error(fmt.Sprintf("failed to get stats for path: %s\n%v"+path, err))
		return false
	}
	return item.IsDir()
}

// Descends a given file tree starting with the given path, assumed to be a directory.
func (c *Client) Discover(dirPath string) (*svc.Directory, error) {
	if !c.isDirPath(dirPath) {
		return nil, fmt.Errorf("path is not a directory: %s", dirPath)
	}

	// see if this directory is already known
	dir, err := c.Db.GetDirectoryByPath(dirPath)
	if err != nil {
		return nil, err
	}
	if dir != nil {
		c.log.Info(fmt.Sprintf("'%s' is already known", filepath.Base(dirPath)))
		return dir, nil
	}

	// create a new directory object and traverse
	c.log.Info(fmt.Sprintf("traversing %s...", dirPath))
	newDir := svc.NewDirectory(filepath.Base(dirPath), c.UserID, c.DriveID, dirPath)
	newDir.Parent = c.Drive.Root
	newDir.BackupPath = filepath.Join(c.Conf.BackupDir, newDir.Name)
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
		if !c.localBackup() {
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
		if !c.localBackup() {
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
	if err := c.Drive.AddSubDir(c.Drive.RootID, newDir); err != nil {
		return nil, fmt.Errorf("failed to add root to drive instance: %v", err)
	}
	if !c.localBackup() {
		if err := c.RegisterDirectory(newDir); err != nil {
			return nil, err
		}
	}
	if err := c.SaveState(); err != nil {
		c.log.Error(fmt.Sprintf("failed to save state file: %v", err))
	}
	return newDir, nil
}

// make sure all items registered on the client side are also registered
// on the client side. if not, register them with the server.
func (c *Client) RegisterItems() error {
	files := c.Drive.GetFiles()
	for _, file := range files {
		reg, err := c.IsFileRegistered(file)
		if err != nil {
			return err
		}
		if !reg {
			if err := c.RegisterFile(file); err != nil {
				c.log.Error(fmt.Sprintf("failed to register file '%s': %v", file.Name, err))
				return err
			}
			serverPath, err := c.getFileServerPath(file)
			if err != nil {
				return err
			}
			file.ServerPath = serverPath
			if err := c.UpdateFile(file); err != nil {
				return err
			}
		}
	}
	dirs := c.Drive.GetDirs()
	for _, dir := range dirs {
		reg, err := c.IsDirRegistered(dir)
		if err != nil {
			return err
		}
		if !reg {
			if err := c.RegisterDirectory(dir); err != nil {
				c.log.Error(fmt.Sprintf("failed to register directory '%s': %v", dir.Name, err))
				return err
			}
			serverPath, err := c.getDirServerPath(dir)
			if err != nil {
				return err
			}
			dir.ServerPath = serverPath
			if err := c.UpdateDirectory(dir); err != nil {
				return err
			}
		}
	}
	return nil
}
