package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/sfs/pkg/auth"
	svc "github.com/sfs/pkg/service"
)

/*
This file contains various implementations of specific server API calls,
for example retrieving a single file or directory, or sending new files
or directories to the server.const

These are basically one-off functions that don't need to be part of
the bigger push/pull synchronization events.
*/

// request information from the server about a particular file, directory
// drive, or user. retrieves only metadata. requires the
func (c *Client) InfoReq(endpoint string, reqType string) error {
	req, err := c.GetInfoRequest(endpoint)
	if err != nil {
		return fmt.Errorf("failed to execute info request: %v", err)
	}
	defer req.Body.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, req.Body)
	if err != nil {
		return err
	}

	// TODO: fancy output

	return nil
}

// executes a file request. can be used to send a new file,
// send an updated file, retrieve a file, or a delete a file on the server.
func (c *Client) FileReq(file *svc.File, reqType string) error {
	req, err := c.GetFileReq(file, reqType)
	if err != nil {
		return err
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		c.dump(resp, true)
		return fmt.Errorf("failed to get file from API: %v", resp.Status)
	}
	defer resp.Body.Close()

	// all other requests involve sending a file. this handles the get case.
	if reqType == "get" {
		// download the file
		var buf bytes.Buffer
		_, err = io.Copy(&buf, req.Body)
		if err != nil {
			return fmt.Errorf("failed to download file: %v", err)
		}
		if err := file.Save(buf.Bytes()); err != nil {
			return fmt.Errorf("failed to save file data: %v", err)
		}
	}
	return nil
}

// executes a directory request API call.
func (c *Client) DirReq(dir *svc.Directory, reqType string) error {
	req, err := c.GetDirReq(dir, reqType)
	if err != nil {
		return err
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		c.dump(resp, true)
		return fmt.Errorf("failed to retrieve directory from API: %v", resp.Status)
	}
	defer resp.Body.Close()

	if reqType == "get" {
		// download the directory as a .zip file and unpack in dest directory.
		var buf bytes.Buffer
		_, err = io.Copy(&buf, req.Body)
		if err != nil {
			return fmt.Errorf("failed to download directory: %v", err)
		}
		// create a tmp file object so we can write out the file data
		tmp := svc.NewFile(dir.Name, dir.DriveID, dir.OwnerID, dir.Path)
		if err := tmp.Save(buf.Bytes()); err != nil {
			return fmt.Errorf("failed to save .zip archive: %v", err)
		}
		// unpack zip file
	}
	return nil
}

func (c *Client) DriveReq(drv *svc.Drive, reqType string) error {
	req, err := c.GetDriveReq(drv, reqType)
	if err != nil {
		return err
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		c.dump(resp, true)
		return fmt.Errorf("failed to retrieve drive from API: %v", resp.Status)
	}
	defer resp.Body.Close()

	if reqType == "get" {
		var buf bytes.Buffer
		_, err = io.Copy(&buf, req.Body)
		if err != nil {
			return fmt.Errorf("failed to download directory: %v", err)
		}
		// TODO: download the drive and unmarshall to drive instance?

	}
	return nil
}

func (c *Client) UserReq(user *auth.User, reqType string) error {
	req, err := c.GetUserReq(user, reqType)
	if err != nil {
		return err
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		c.dump(resp, true)
		return fmt.Errorf("failed to retrieve user info from API: %v", resp.Status)
	}
	defer resp.Body.Close()

	if reqType == "get" {
		var buf bytes.Buffer
		_, err = io.Copy(&buf, req.Body)
		if err != nil {
			return fmt.Errorf("failed to download user info: %v", err)
		}
		// TODO: download the user info and unmarshall to user instance?

	}
	return nil
}
