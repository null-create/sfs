package client

import (
	"fmt"
	"net/http"

	svc "github.com/sfs/pkg/service"
)

/*
File for coordinating file/directory transfers between the client and the server
May need to build SyncIndex.ToUpdate map, or transfer individual files or directories

Some functions may need to be connected to the client struct. Maybe not. We'll see.

Will probably need to create some functions that "fan-out" or "fan-in" goroutines,
depending on the number of items to push or pull from the server.

Will also probably need to make use of the batches, queues, and some functions
(like BuildQ()) defined in sync.go within in the core service module.
*/

// take a given synch index, build a queue of files to be pushed to the
// server, then upload each in their own goroutines
func (c *Client) Push() error {
	if len(c.Drive.SyncIndex.ToUpdate) == 0 {
		return fmt.Errorf("no files marked for uploading. SyncIndex.ToUpdate is empty")
	}
	// build file/dir queue for uploading
	queue := svc.BuildQ(c.Drive.SyncIndex)
	if len(queue.Queue) == 0 || queue == nil { // this might be a bit redundant
		return fmt.Errorf("unable to build queue: no files found for syncing")
	}
	// 'fan-out' individual upload goroutines to the server
	for _, batch := range queue.Queue {
		for _, file := range batch.Files {
			// TODO: replace file.ServerPath with destURL/server API for this file
			go c.Transfer.Upload(http.MethodPost, file, file.ServerPath)
		}
	}
	return nil
}

// get a sync index from the server, compare with the local one
// with the client, and pull any files that are out of date on the client side
// create goroutines for each download and 'fan-in' once all are complete
func (c *Client) Pull() error { return nil }

// background daemon that listens for requests from the server to
// download files. returns a channel that is used to shut down the daemon
// if needed
func (c *Client) ListenerDaemon() (chan bool, error) {
	// 1. establish connection with server
	// 2. start a blocking net listener for pings from the server
	// 3. if we get a request, get the request type (get from server or push to server),
	//    then generate a list of files or directories to be sent to the server or pulled from
	//    the server
	return nil, nil
}
