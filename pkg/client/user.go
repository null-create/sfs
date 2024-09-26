package client

import (
	"fmt"

	"github.com/sfs/pkg/auth"
)

// Add a new user. Intended for use during first time set up.
func (c *Client) AddNewUser() error {
	if c.User != nil {
		fmt.Printf("User '%s' already exists", c.User.Name)
		return nil
	}

	var name, userName, password, email string
	c.getInput("Name: ", name)
	c.getInput("User name: ", userName)
	c.getInput("Email: ", email)
	c.getInput("Password (leave blank to auto-generate): ", password)
	if password == "" {
		password = auth.GenSecret(64)
	}

	newUser := auth.NewUser(name, userName, email, cCfgs.Root, false)
	newUser.Password = password
	if err := c.Db.AddUser(newUser); err != nil {
		return err
	}

	var newUserSettings = map[string]string{
		"CLIENT_NAME":     newUser.Name,
		"CLIENT_USERNAME": newUser.UserName,
		"CLIENT_EMAIL":    newUser.Email,
		"CLIENT_PASSWORD": newUser.Password,
		"CLIENT_ID":       newUser.ID,
	}
	for setting, value := range newUserSettings {
		if err := c.UpdateConfigSetting(setting, value); err != nil {
			return err
		}
	}

	if err := c.SaveState(); err != nil {
		return err
	}
	return nil
}

func (c *Client) LoadUser() error {
	if c.User == nil {
		if c.UserID == "" {
			return fmt.Errorf("missing user id")
		}
		user, err := c.Db.GetUser(c.UserID)
		if err != nil {
			return fmt.Errorf("failed to get user (id=%v): ", err)
		} else if user == nil {
			return fmt.Errorf("user (id=%s) not found", c.UserID)
		}
		c.User = user
		if err := c.SaveState(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) GetUserInfo() string {
	if c.User == nil {
		c.log.Error("user not found")
		return ""
	}
	data, err := c.User.ToJSON()
	if err != nil {
		c.log.Error(fmt.Sprintf("error getting user info: %v", err))
		return ""
	}
	return string(data)
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
func (c *Client) RemoveUser() error {
	if c.User == nil {
		return fmt.Errorf("no user found")
	}
	if err := c.removeUser(c.User.ID); err != nil {
		return err
	}
	return nil
}

func (c *Client) removeUser(userID string) error {
	fmt.Printf("user '%s' and all their monitored files and directories will be removed from SFS", c.User.Name)
	if c.Continue() {
		files := c.Drive.GetFiles()
		if err := c.Db.RemoveFiles(files); err != nil {
			return err
		}
		dirs := c.Drive.GetDirs()
		if err := c.Db.RemoveDirectories(dirs); err != nil {
			return err
		}
		if err := c.Drive.ClearDrive(); err != nil {
			return err
		}
		if err := c.Db.RemoveDrive(c.Drive.ID); err != nil {
			return err
		}
		if err := c.Db.RemoveUser(c.UserID); err != nil {
			return err
		}
		if err := svcCfgs.Clear(); err != nil {
			return err
		}
		name := c.User.Name
		c.User = nil
		c.log.Info(fmt.Sprintf("'%s' (id=%s) removed", name, userID))
		if err := c.SaveState(); err != nil {
			return err
		}
	}
	return nil
}
