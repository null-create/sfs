package client

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/sfs/pkg/auth"
	svc "github.com/sfs/pkg/service"
)

/*
File for constructing and sending requests to the server API.

Will primarily use the client's built in HTTP client for this.
Use Models package to help create requests.
*/

// utils for helping create specific http request objects

// ------- token creation --------------------------------

// createa new jwt token with the given payload
func (c *Client) NewToken(payload string) (string, error) {
	reqToken, err := c.Tok.Create(payload)
	if err != nil {
		return "", fmt.Errorf("failed to create token: %v", err)
	}
	return reqToken, nil
}

func (c *Client) encodeUser(user *auth.User) (string, error) {
	payload, err := user.ToJSON()
	if err != nil {
		return "", err
	}
	return c.NewToken(string(payload))
}

func (c *Client) encodeFile(file *svc.File) (string, error) {
	payload, err := file.ToJSON()
	if err != nil {
		return "", err
	}
	return c.NewToken(string(payload))
}

func (c *Client) encodeDir(dir *svc.Directory) (string, error) {
	payload, err := dir.ToJSON()
	if err != nil {
		return "", err
	}
	return c.NewToken(string(payload))
}

func (c *Client) encodeDrive(drv *svc.Drive) (string, error) {
	payload, err := drv.ToJSON()
	if err != nil {
		return "", err
	}
	return c.NewToken(string(payload))
}

// ------- get request objects --------------------------------------------------

// valid reqTypes: "new", "get", "update", "delete"
func (c *Client) GetFileReq(file *svc.File, reqType string) (*http.Request, error) {
	switch reqType {
	case "new":
		return c.NewFileRequest(file)
	case "get":
		return c.GetFileRequest(file)
	case "update":
		return c.UpdateFileRequest(file)
	case "delete":
		return c.DeleteFileRequest(file)
	default:
		return nil, fmt.Errorf("unsupported request type")
	}
}

// valid reqTypes: "new", "get", "update", "delete"
func (c *Client) GetDirReq(dir *svc.Directory, reqType string) (*http.Request, error) {
	switch reqType {
	case "new":
		return c.NewDirectoryRequest(dir)
	case "get":
		return c.GetDirRequest(dir)
	case "update":
		return c.UpdateDirectoryRequest(dir)
	case "delete":
		return c.DeleteDirectoryRequest(dir)
	default:
		return nil, fmt.Errorf("unsupported request type")
	}
}

// valid reqTypes: "new", "get", "update", "delete"
func (c *Client) GetDriveReq(drv *svc.Drive, reqType string) (*http.Request, error) {
	switch reqType {
	case "new":
		return c.NewDriveRequest(drv)
	case "get":
		return c.GetDriveRequest(drv)
	case "update":
		return c.UpdateDriveRequest(drv)
	case "delete":
		return c.DeleteDriveRequest(drv)
	default:
		return nil, fmt.Errorf("unsupported request type")
	}
}

// valid reqTypes: "new", "get", "update", "delete"
func (c *Client) GetUserReq(user *auth.User, reqType string) (*http.Request, error) {
	switch reqType {
	case "new":
		return c.NewUserRequest(user)
	case "get":
		return c.GetUserRequest(user)
	case "update":
		return c.UpdateUserRequest(user)
	case "delete":
		return c.DeleteUserRequest(user)
	default:
		return nil, fmt.Errorf("unsupported request type")
	}
}

// ------ new item requests ----------------------------------------------

func (c *Client) NewUserRequest(newUser *auth.User) (*http.Request, error) {
	var buf bytes.Buffer
	req, err := http.NewRequest(http.MethodPost, c.Endpoints["new user"], &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	reqToken, err := c.encodeUser(newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create request token: %v", err)
	}
	req.Header.Set("Authorization", reqToken)
	return req, nil
}

func (c *Client) NewFileRequest(newFile *svc.File) (*http.Request, error) {
	var buf bytes.Buffer
	req, err := http.NewRequest(http.MethodPost, c.Endpoints["new file"], &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	reqToken, err := c.encodeFile(newFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create request token: %v", err)
	}
	req.Header.Set("Authorization", reqToken)
	return req, nil
}

func (c *Client) NewDirectoryRequest(newDir *svc.Directory) (*http.Request, error) {
	var buf bytes.Buffer
	req, err := http.NewRequest(http.MethodPost, c.Endpoints["new dir"], &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	reqToken, err := c.encodeDir(newDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create request token: %v", err)
	}
	req.Header.Set("Authorization", reqToken)
	return req, nil
}

func (c *Client) NewDriveRequest(newDrv *svc.Drive) (*http.Request, error) {
	var buf bytes.Buffer
	req, err := http.NewRequest(http.MethodPost, c.Endpoints["new drive"], &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	reqToken, err := c.encodeDrive(newDrv)
	if err != nil {
		return nil, fmt.Errorf("failed to create request token: %v", err)
	}
	req.Header.Set("Authorization", reqToken)
	return req, nil
}

// ----- get

func (c *Client) GetFileRequest(file *svc.File) (*http.Request, error) {
	var buf bytes.Buffer
	req, err := http.NewRequest(http.MethodGet, file.Endpoint, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	reqToken, err := c.encodeFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create request token: %v", err)
	}
	req.Header.Set("Authorization", reqToken)
	return req, nil
}

func (c *Client) GetDirRequest(dir *svc.Directory) (*http.Request, error) {
	var buf bytes.Buffer
	req, err := http.NewRequest(http.MethodGet, dir.Endpoint, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	reqToken, err := c.encodeDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to create request token: %v", err)
	}
	req.Header.Set("Authorization", reqToken)
	return req, nil
}

func (c *Client) GetDriveRequest(drv *svc.Drive) (*http.Request, error) {
	var buf bytes.Buffer
	req, err := http.NewRequest(http.MethodGet, c.Endpoints["drive"], &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	reqToken, err := c.encodeDrive(drv)
	if err != nil {
		return nil, fmt.Errorf("failed to create request token: %v", err)
	}
	req.Header.Set("Authorization", reqToken)
	return req, nil
}

func (c *Client) GetUserRequest(user *auth.User) (*http.Request, error) {
	var buf bytes.Buffer
	req, err := http.NewRequest(http.MethodGet, c.Endpoints["user"], &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	reqToken, err := c.encodeUser(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create request token: %v", err)
	}
	req.Header.Set("Authorization", reqToken)
	return req, nil
}

// ----- update

func (c *Client) UpdateFileRequest(file *svc.File) (*http.Request, error) {
	var buf bytes.Buffer
	req, err := http.NewRequest(http.MethodPut, file.Endpoint, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	reqToken, err := c.encodeFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create request token: %v", err)
	}
	req.Header.Set("Authorization", reqToken)
	return req, nil
}

func (c *Client) UpdateDirectoryRequest(dir *svc.Directory) (*http.Request, error) {
	var buf bytes.Buffer
	req, err := http.NewRequest(http.MethodPut, dir.Endpoint, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	reqToken, err := c.encodeDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to create request token: %v", err)
	}
	req.Header.Set("Authorization", reqToken)
	return req, nil
}

func (c *Client) UpdateDriveRequest(drv *svc.Drive) (*http.Request, error) {
	var buf bytes.Buffer
	req, err := http.NewRequest(http.MethodPut, c.Endpoints["drive"], &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	reqToken, err := c.encodeDrive(drv)
	if err != nil {
		return nil, fmt.Errorf("failed to create request token: %v", err)
	}
	req.Header.Set("Authorization", reqToken)
	return req, nil
}

func (c *Client) UpdateUserRequest(user *auth.User) (*http.Request, error) {
	var buf bytes.Buffer
	req, err := http.NewRequest(http.MethodPut, c.Endpoints["user"], &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	reqToken, err := c.encodeUser(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create request token: %v", err)
	}
	req.Header.Set("Authorization", reqToken)
	return req, nil
}

// ---- delete

func (c *Client) DeleteFileRequest(file *svc.File) (*http.Request, error) {
	var buf bytes.Buffer
	req, err := http.NewRequest(http.MethodDelete, file.Endpoint, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	reqToken, err := c.encodeFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create request token: %v", err)
	}
	req.Header.Set("Authorization", reqToken)
	return req, nil
}

func (c *Client) DeleteDirectoryRequest(dir *svc.Directory) (*http.Request, error) {
	var buf bytes.Buffer
	req, err := http.NewRequest(http.MethodDelete, dir.Endpoint, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	reqToken, err := c.encodeDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to create request token: %v", err)
	}
	req.Header.Set("Authorization", reqToken)
	return req, nil
}

func (c *Client) DeleteDriveRequest(drv *svc.Drive) (*http.Request, error) {
	var buf bytes.Buffer
	req, err := http.NewRequest(http.MethodDelete, c.Endpoints["drive"], &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	reqToken, err := c.encodeDrive(drv)
	if err != nil {
		return nil, fmt.Errorf("failed to create request token: %v", err)
	}
	req.Header.Set("Authorization", reqToken)
	return req, nil
}

func (c *Client) DeleteUserRequest(user *auth.User) (*http.Request, error) {
	var buf bytes.Buffer
	req, err := http.NewRequest(http.MethodDelete, c.Endpoints["user"], &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	reqToken, err := c.encodeUser(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create request token: %v", err)
	}
	req.Header.Set("Authorization", reqToken)
	return req, nil
}
