package client

/*
File for coordinating file/directory transfers between the client and the server
May need to build SyncIndex.ToUpdate map, or transfer individual files or directories
*/

import (
	"fmt"
	"log"
	"net/http"

	svc "github.com/sfs/pkg/service"
)

// take a given synch index, build a queue of files to be pushed to the
// server, then upload each in their own goroutines
func (c *Client) Push() error {
	if len(c.Drive.SyncIndex.ToUpdate) == 0 {
		return fmt.Errorf("no files marked for uploading. SyncIndex.ToUpdate is empty")
	}
	// build file/dir queue for uploading
	queue := svc.BuildQ(c.Drive.SyncIndex)
	if len(queue.Queue) == 0 || queue == nil { // the "or" might be a bit redundant
		return fmt.Errorf("unable to build queue: no files found for syncing")
	}
	// 'fan-out' individual upload goroutines to the server
	for _, batch := range queue.Queue {
		for _, file := range batch.Files {
			// TODO: some apis are contingent on http method: file post/put is new vs update
			// need a way to handle these cases on the fly.
			// maybe add a field to File struct?
			go func() {
				if err := c.Transfer.Upload(http.MethodPost, file, file.Endpoint); err != nil {
					log.Printf("[WARNING] failed to upload file: %s\nerr: %v", file.ID, err)
				}
			}()
		}
	}
	return nil
}

// get a sync index from the server, compare with the local one
// with the client, and pull any files that are out of date on the client side
// create goroutines for each download and 'fan-in' once all are complete
func (c *Client) Pull(svrIdx *svc.SyncIndex) error {
	if svrIdx == nil || len(svrIdx.ToUpdate) == 0 {
		log.Print("[INFO] nothing to pull")
		return nil
	}
	// build file/dir queue for downloading
	queue := svc.BuildQ(svrIdx)
	if len(queue.Queue) == 0 || queue == nil { // the "or" might be a bit redundant
		return fmt.Errorf("unable to build queue: no files found for syncing")
	}
	for _, batch := range queue.Queue {
		for _, file := range batch.Files {
			go func() {
				if err := c.Transfer.Download(file.ServerPath, file.Endpoint); err != nil {
					log.Printf("[WARNING] failed to download file: %s\nerr: %v", file.Name, err)
				}
			}()
		}
	}
	return nil
}
