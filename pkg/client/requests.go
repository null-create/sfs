package client

import (
	"github.com/sfs/pkg/auth"
	svc "github.com/sfs/pkg/service"
)

/*
File for constructing and sending requests to the server

Will primarily use the client's built in HTTP client for this.
Use Models package to help create requests.
*/

// utils for helping create specific requests

// creation

func (c *Client) NewUserRequest(newUser *auth.User) error { return nil }

func (c *Client) NewFileRequest(newFile *svc.File) error { return nil }

func (c *Client) NewDirectoryRequest(newDir *svc.Directory) error { return nil }

func (c *Client) NewDriveRequest(newDrv *svc.Drive) error { return nil }

// updates/deletes

func (c *Client) UpdateFileRequest(file *svc.File) error { return nil }

func (c *Client) DeleteFileRequest(file *svc.File) error { return nil }

func (c *Client) UpdateDirectoryRequest(dir *svc.Directory) error { return nil }

func (c *Client) DeleteDirectoryRequest(dir *svc.Directory) error { return nil }

func (c *Client) UpdateDriveRequest(drv *svc.Drive) error { return nil }

func (c *Client) DeleteDriveRequest(drv *svc.Drive) error { return nil }
