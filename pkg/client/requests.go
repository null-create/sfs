package client

/*
File for constructing and sending requests to the server

Will primarily use the client's built in HTTP client for this.
Use Models package to help create requests.
*/

// utils for helping create specific requests

// creation

func (c *Client) NewUserRequest() error { return nil }

func (c *Client) NewFileRequest() error { return nil }

func (c *Client) NewDirectoryRequest() error { return nil }

func (c *Client) NewDriveRequest() error { return nil }

// updates/deletes

func (c *Client) UpdateFileRequest() error { return nil }

func (c *Client) DeleteFileRequest() error { return nil }

func (c *Client) UpdateDirectoryRequest() error { return nil }

func (c *Client) DeleteDirectoryRequest() error { return nil }

func (c *Client) UpdateDriveRequest() error { return nil }

func (c *Client) DeleteDriveRequest() error { return nil }
