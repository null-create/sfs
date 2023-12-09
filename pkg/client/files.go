package client

import (
	"fmt"
	"log"

	"github.com/sfs/pkg/auth"
	svc "github.com/sfs/pkg/service"
)

// ------- user functions --------------------------------

func (c *Client) AddUser(user *auth.User) error {
	if c.User == nil {
		c.User = user
	} else {
		return fmt.Errorf("cannot have more than one user: %v", c.User)
	}
	if err := c.SaveState(); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetUser() (*auth.User, error) {
	if c.User == nil {
		log.Print("[WARNING] client instance has no user object. attempting to get user from the database...")
		if c.Db == nil {
			return nil, fmt.Errorf("failed to get user. database not initialized")
		}
		user, err := c.Db.GetUser(c.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user from database: %v", err)
		}
		return user, nil
	}
	return c.User, nil
}

func (c *Client) UpdateUser(user *auth.User) error {
	if user.ID == c.User.ID {
		c.User = user
	} else {
		return fmt.Errorf("user (id=%s) is not client user (id=%s)", user.ID, c.User.ID)
	}
	if err := c.SaveState(); err != nil {
		return err
	}
	return nil
}

// remove a user and their drive from the client instance.
// clears all users files and directores, as well as removes the
// user from the client instance and db.
func (c *Client) RemoveUser(userID string) error {
	if c.User == nil {
		return fmt.Errorf("no user (id=%s) found", userID)
	} else if c.User.ID == userID {
		// remove drive and users files
		c.Drive.Remove()
		// remove user and info from database
		if err := c.Db.RemoveUser(c.UserID); err != nil {
			return err
		}
		c.User = nil
		log.Printf("[INFO] user %s removed", userID)
	} else {
		return fmt.Errorf("wrong user ID (id=%s)", userID)
	}
	if err := c.SaveState(); err != nil {
		return err
	}
	return nil
}

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

func (c *Client) AddDir(dir *svc.Directory) error {
	return c.Drive.AddDir(dir)
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
