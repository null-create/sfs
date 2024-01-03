package client

/*
File for coordinating file/directory transfers between the client and the server
May need to build SyncIndex.ToUpdate map, or transfer individual files or directories
*/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	svc "github.com/sfs/pkg/service"
)

const (
	EndpointRoot = "http://localhost"
	CheckWait    = time.Millisecond * 500
	SyncWait     = time.Minute
)

// gets server Sync index, compares with local index, and either
// calls Push or Pull, depending on whether the corresponding bool
// flag is set. resets Sync doc too.
func (c *Client) Sync(up bool) error {
	// TODO: compare with local index, find any differences between them,
	// and determine whether to push or pull (or both, if the server has
	// some newer versions of local files, and the client has newer (or new)
	// files to update/send to the server)
	if up {
		localIdx := c.Drive.SyncIndex
		if localIdx == nil {
			if !c.Drive.HasRoot() {
				return fmt.Errorf("drive root has not been instantiated")
			}
			localIdx = svc.BuildSyncIndex(c.Drive.Root)
			c.Drive.SyncIndex = svc.BuildToUpdate(c.Drive.Root, localIdx)
		}
		c.Push()
	} else {
		svrIdx := c.GetServerIdx()
		if svrIdx == nil {
			return fmt.Errorf("failed to retrieve server sync index")
		}
		c.Pull(svrIdx)
	}
	c.reset()
	return nil
}

// resets client side sync mechanisms
func (c *Client) reset() {
	c.Drive.SyncIndex.Reset() // clear ToUpdate
}

// display server response clearly.
func (c *Client) dump(resp *http.Response, body bool) {
	b, err := httputil.DumpResponse(resp, body)
	if err != nil {
		log.Printf("[WARNING] failed to dump http response:\n%v", err)
	} else {
		log.Printf("\n%s\n", string(b))
	}
}

// take a given synch index, build a queue of files to be pushed to the
// server, then upload each in their own goroutines

// TODO: handle the difference between creates and updates.
// some files may be new, others may be only modified!
func (c *Client) Push() error {
	if len(c.Drive.SyncIndex.ToUpdate) == 0 || c.Drive.SyncIndex.ToUpdate == nil {
		return fmt.Errorf("no files marked for uploading. SyncIndex.ToUpdate is empty")
	}
	queue := svc.BuildQ(c.Drive.SyncIndex)
	if len(queue.Queue) == 0 || queue == nil {
		return fmt.Errorf("unable to build queue: no files found for syncing")
	}
	// TODO: use a channel to block Push() right before c.reset() until
	// all files have been uploaded
	// for _, batch := range queue.Queue {
	for len(queue.Queue) > 0 {
		batch := queue.Dequeue()
		for _, file := range batch.Files {
			// TODO: some apis are contingent on http method: file post/put is new vs update
			// need a way to handle these cases on the fly.
			go func() {
				log.Printf("[INFO] uploading %s to %s...", file.Name, file.Endpoint)
				if err := c.Transfer.Upload(
					http.MethodPost,
					file,
					file.Endpoint,
				); err != nil {
					log.Printf("[WARNING] failed to upload file: %s\nerr: %v", file.ID, err)
				}
			}()
		}
	}
	c.reset()
	return nil
}

// gets a sync index from the server, compares with the local one,
// and pulls any files that are out of date on the client side from the server.
// create goroutines for each download and 'fans-in' once all are complete.
func (c *Client) Pull(svrIdx *svc.SyncIndex) error {
	if len(svrIdx.ToUpdate) == 0 {
		log.Print("[INFO] no sync index return from server. nothing to pull")
		return nil
	}
	queue := svc.BuildQ(svrIdx)
	if len(queue.Queue) == 0 || queue == nil {
		return fmt.Errorf("unable to build queue: no files found for syncing")
	}
	for _, batch := range queue.Queue {
		for _, file := range batch.Files {
			go func() {
				if err := c.Transfer.Download(
					file.ClientPath,
					file.Endpoint,
				); err != nil {
					log.Printf("[WARNING] failed to download file: %s\nerr: %v", file.Name, err)
				}
				if err := file.ValidateChecksum(); err != nil {
					log.Printf("[WARNING] failed to validate checksum for file %v", file.Name)
				}
				if err := c.Db.UpdateFile(file); err != nil {
					log.Printf("[ERROR] failed to update files database: %v", err)
				}
			}()
		}
	}
	c.reset()
	return nil
}

// get the server's current sync index for this user.
// returns nil if there's any errors.
func (c *Client) GetServerIdx() *svc.SyncIndex {
	req, err := http.NewRequest(http.MethodGet, c.Endpoints["sync"], new(bytes.Buffer))
	if err != nil {
		log.Printf("[WARNING] failed prepare http request: \n%v", err)
		return nil
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		log.Printf("[WARNING] failed to get execute http request: \n%v", err)
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("[WARNING] failed to get server sync index. return code: %d", resp.StatusCode)
		c.dump(resp, true)
		return nil
	}

	data := make([]byte, 0)
	_, err = resp.Body.Read(data)
	if err != nil {
		log.Printf("[WARNING] failed to read server response body: \n%v", err)
		return nil
	}
	defer resp.Body.Close()

	idx := new(svc.SyncIndex)
	if err := json.Unmarshal(data, &idx); err != nil {
		log.Printf("[WARNING] failed to unmarshal server sync index: \n%v", err)
		return nil
	}
	return idx
}
