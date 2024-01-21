package client

import (
	"fmt"
	"log"

	"github.com/sfs/pkg/auth"
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
		if err := c.Drive.ClearDrive(); err != nil {
			return err
		}
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
