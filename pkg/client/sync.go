package client

/*
File for coordinating file/directory transfers between the client and the server
May need to build SyncIndex.ToUpdate map, or transfer individual files or directories
*/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"

	svc "github.com/sfs/pkg/service"
)

const (
	EndpointRoot = "http://localhost"
	CheckWait    = time.Millisecond * 500
	SyncWait     = time.Minute
)

// resets client side sync mechanisms
func (c *Client) reset() {
	c.Drive.SyncIndex.Reset()
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

// TODO: handle the difference between creates and updates.
// some files may be new, others may be only modified!
//
// take a given synch index, build a queue of files to be pushed to the
// server, then upload each in their own goroutines
func (c *Client) Push() error {
	if len(c.Drive.SyncIndex.ToUpdate) == 0 {
		return fmt.Errorf("no files marked for uploading. SyncIndex.ToUpdate is empty")
	}
	queue := svc.BuildQ(c.Drive.SyncIndex)
	if queue == nil {
		return fmt.Errorf("unable to build queue: no files found for syncing")
	}
	var wg sync.WaitGroup
	for len(queue.Queue) > 0 {
		batch := queue.Dequeue()
		for _, file := range batch.Files {
			// TODO: some apis are contingent on http method: file post/put is new vs update
			// need a way to handle these cases on the fly.
			wg.Add(1)
			go func() {
				defer wg.Done()
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
	wg.Wait()
	c.reset()
	return nil
}

// gets a sync index from the server, compares with the local one,
// and pulls any files that are out of date on the client side from the server.
// create goroutines for each download and 'fans-in' once all are complete.
func (c *Client) Pull(idx *svc.SyncIndex) error {
	if len(idx.ToUpdate) == 0 {
		log.Print("[INFO] no sync index return from server. nothing to pull")
		return nil
	}
	queue := svc.BuildQ(idx)
	if len(queue.Queue) == 0 || queue == nil {
		return fmt.Errorf("unable to build queue: no files found for syncing")
	}
	var wg sync.WaitGroup
	for len(queue.Queue) > 0 {
		batch := queue.Dequeue()
		for _, file := range batch.Files {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := c.Transfer.Download(
					file.ClientPath,
					file.Endpoint,
				); err != nil {
					log.Printf("[WARNING] failed to download file: %s\nerr: %v", file.Name, err)
					return
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
	wg.Wait()
	c.reset()
	return nil
}

// TODO:
func (c *Client) Diff() error {
	// refresh ToUpdate
	if !c.Drive.IsIndexed() {
		return fmt.Errorf("no files found for indexing")
	}
	c.Drive.SyncIndex = svc.BuildToUpdate(c.Drive.Root, c.Drive.SyncIndex)

	// retrieve the servers index for this client
	resp, err := c.Client.Get(c.Endpoints["gen updates"])
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] server returned status code: %v", resp.StatusCode)
		c.dump(resp, true)
		return nil
	}

	// TODO: compare the two indicies, and generate a third index
	// with the most recent LastSync times for each file and directory,
	// then display the differences.
	return nil
}

// TODO:
func (c *Client) Sync() error {
	// get latest local index/Update map
	if !c.Drive.IsIndexed() {
		return fmt.Errorf("no files found for indexing")
	}
	c.Drive.SyncIndex = svc.BuildToUpdate(c.Drive.Root, c.Drive.SyncIndex)

	// get latest server update map
	resp, err := c.Client.Get(c.Endpoints["gen updates"])
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("[INFO] server returned non-200 status")
		c.dump(resp, true)
		return nil
	}
	defer resp.Body.Close()

	// retrieve index from response body
	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return err
	}
	var svrIdx = new(svc.SyncIndex)
	if err := json.Unmarshal(buf.Bytes(), &svrIdx); err != nil {
		return err
	}

	// compare the two, then push/pull accordingly

	return nil
}

// retrieve the current sync index for this user from the server
func (c *Client) GetServerIdx() (*svc.SyncIndex, error) {
	resp, err := c.Client.Get(c.Endpoints["get index"])
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		c.dump(resp, true)
		return nil, fmt.Errorf("failed to get server sync index: %v", resp.StatusCode)
	}
	defer resp.Body.Close()

	var idxBuf bytes.Buffer
	_, err = io.Copy(&idxBuf, resp.Body)
	if err != nil {
		return nil, err
	}
	var idx *svc.SyncIndex
	if err = json.Unmarshal(idxBuf.Bytes(), &idx); err != nil {
		return nil, err
	}
	return idx, nil
}

// TODO:
// ------- single-operation pushes and pulls from the server -------------
