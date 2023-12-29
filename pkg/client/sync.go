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
	"os"
	"time"

	svc "github.com/sfs/pkg/service"
)

const (
	EndpointRoot = "http://localhost"
	CheckWait    = time.Millisecond * 500
	SyncWait     = time.Minute
)

// gets server sync index, compares with local index, and either
// calls Push or Pull, depending on whether the corresponding bool
// flag is set. resets sync doc too.
func (c *Client) sync(up bool) error {
	if up {
		c.Push()
	} else {
		c.Pull()
	}
	c.reset()
	return nil
}

// resets client side sync mechanisms
func (c *Client) reset() {
	c.Drive.SyncIndex.Reset() // clear ToUpdate
	c.Monitor.ResetDoc()      // resets sync doc
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
	for _, batch := range queue.Queue {
		for _, file := range batch.Files {
			// TODO: some apis are contingent on http method: file post/put is new vs update
			// need a way to handle these cases on the fly.
			go func() {
				if err := c.Transfer.Upload(http.MethodPost, file, file.Endpoint); err != nil {
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
func (c *Client) Pull() error {
	svrIdx := c.GetServerIdx()
	if svrIdx == nil {
		log.Print("[WARNING] no sync index available from server. unable to pull files.")
		return nil
	}
	if len(svrIdx.ToUpdate) == 0 || svrIdx == nil {
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
				// download file
				if err := c.Transfer.Download(file.ClientPath, file.Endpoint); err != nil {
					log.Printf("[WARNING] failed to download file: %s\nerr: %v", file.Name, err)
				}
				// validate checksum
				if err := file.ValidateChecksum(); err != nil {
					log.Printf("[WARNING] failed to validate checksum for file %v", file.Name)
				}
				// update files DB
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
	// make a new request
	buffer := new(bytes.Buffer)
	req, err := http.NewRequest(http.MethodGet, c.Endpoints["sync"], buffer)
	if err != nil {
		log.Printf("[WARNING] failed prepare http request: \n%v", err)
		return nil
	}

	// attempt to get the server's sync index
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

	// parse response body
	respBody := make([]byte, resp.ContentLength)
	_, err = resp.Body.Read(respBody)
	if err != nil {
		log.Printf("[WARNING] failed to read server response body: \n%v", err)
		return nil
	}
	defer resp.Body.Close()

	// unmarshal response body
	idx := new(svc.SyncIndex)
	if err := json.Unmarshal(respBody, &idx); err != nil {
		log.Printf("[WARNING] failed to unmarshal server sync index: \n%v", err)
		return nil
	}
	return idx
}

// periodically checks sync doc for whether a sync
// operation should be performed
func (c *Client) Sync(stop chan bool) {
	log.Print("[INFO] starting sync monitoring thread...")
	go func() {
		for {
			select {
			case <-stop:
				log.Print("[INFO] stopping sync thread...")
				return
			default:
				contents, err := os.ReadFile(c.Monitor.SyncDoc)
				if err != nil {
					log.Printf("[WARNING] failed to read sync doc: %v\nstopping sync thread...", err)
					return
				}
				if string(contents) == "1" {
					// TODO: determine whether we should push or pull at sync time.
					// currently defaults to push.
					c.sync(true)
					time.Sleep(SyncWait)
				} else {
					time.Sleep(CheckWait)
				}
			}
		}
	}()
}
