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
	"net/http"
	"net/http/httputil"
	"sync"
	"time"

	svc "github.com/sfs/pkg/service"
)

const (
	EndpointRoot = "http://localhost"
	CheckWait    = time.Millisecond * 500
	SyncWait     = time.Second * 30
)

// whether auto sync is enabled
func (c *Client) autoSync() bool { return c.Conf.AutoSync }

// resets client side sync mechanisms with a
// new baseline for item last sync times.
func (c *Client) reset() {
	c.Drive.SyncIndex.Reset()
	c.Drive.SyncIndex = svc.BuildSyncIndex(c.Drive.Root)
}

// display server response clearly.
func (c *Client) dump(resp *http.Response, body bool) {
	b, err := httputil.DumpResponse(resp, body)
	if err != nil {
		c.log.Warn("failed to dump http response: " + err.Error())
	} else {
		if resp.StatusCode != http.StatusOK {
			c.log.Warn("request failed: " + resp.Status)
			c.log.Log("WARN", "server response: \n"+string(b))
		} else {
			c.log.Show("server response: " + resp.Status)
		}
	}
}

// enable or disable auto sync with the server.
func (c *Client) SetAutoSync(mode bool) {
	c.Conf.AutoSync = mode
	if err := c.SaveState(); err != nil {
		c.log.Error("failed to update state file: " + err.Error())
	} else {
		if mode {
			c.log.Info("auto sync enabled")
		} else {
			c.log.Info("auto sync disabled")
		}
	}
}

type SyncItems struct {
	pull []*svc.File
	push []*svc.File
}

// sync items between the client and the server.
//
// NOTE: this assumes that both the client and the server have
// a record of the files. if the server has a file the client doesn't
// know about, then this doesn't handle it, and vice-versa
//
// TODO: handle when a server has a file/directory the client doesn't have,
// and handle when the client has a file/directory the server doesn't have.
func (c *Client) Sync() error {
	// get latest server sync index
	svrIdx, err := c.GetServerIdx(true)
	if err != nil {
		return err
	}

	var syncItems = new(SyncItems)
	var localIndex = c.Drive.SyncIndex

	// figure out which items to push and pull
	for id, svrLastSync := range svrIdx.LastSync {
		if localIndex.HasItem(id) {
			if svrLastSync.After(localIndex.LastSync[id]) {
				file, err := c.GetFileByID(id)
				if err != nil {
					return err
				}
				syncItems.pull = append(syncItems.pull, file)
			} else if localIndex.LastSync[id].After(svrIdx.LastSync[id]) {
				file, err := c.GetFileByID(id)
				if err != nil {
					return err
				}
				syncItems.push = append(syncItems.push, file)
			}
		}
	}
	if len(syncItems.pull) == 0 && len(syncItems.push) == 0 {
		c.log.Info("no sync operation necessary. exiting...")
		return nil
	}

	// pull items
	var wg sync.WaitGroup
	c.log.Info(fmt.Sprintf("pulling %d files from the server...", len(syncItems.pull)))
	for _, file := range syncItems.pull {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.PullFile(file); err != nil {
				c.log.Error(fmt.Sprintf("failed to pull file: %v", err))
			}
		}()
	}
	wg.Wait()

	// push items
	c.log.Info(fmt.Sprintf("pushing %d files to the server...", len(syncItems.push)))
	for _, file := range syncItems.push {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.PushFile(file); err != nil {
				c.log.Error("failed to push file: " + err.Error())
			}
		}()
	}
	wg.Wait()

	// reset local sync mechanisms
	c.reset()

	return nil
}

// take a given sync index, build a queue of files to be pushed to the
// server, then upload each in their own goroutines. Each file is assumed to be
// already registered with the server, otherwise this will receive a 404 response
// and the upload will fail.
func (c *Client) Push() error {
	if len(c.Drive.SyncIndex.FilesToUpdate) == 0 {
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
			wg.Add(1)
			go func() {
				defer wg.Done()
				c.log.Info(fmt.Sprintf("Uploading %s...", file.Name))
				if err := c.Transfer.Upload(
					http.MethodPut,
					file,
					file.Endpoint,
				); err != nil {
					c.log.Warn(fmt.Sprintf("failed to upload file (name=%s id=%s): %v", file.Name, file.ID, err))
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
	if len(idx.FilesToUpdate) == 0 {
		c.log.Warn("no sync index returned from the server. nothing to pull")
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
					c.log.Warn(fmt.Sprintf("failed to download %s: %v", file.Name, err))
					return
				}
				if err := file.ValidateChecksum(); err != nil {
					c.log.Warn(fmt.Sprintf("failed to validate checksum for %s: %v", file.Name, err))
				}
				if err := c.Db.UpdateFile(file); err != nil {
					c.log.Warn(fmt.Sprintf("failed to update files database: %v", err))
				}
			}()
		}
	}
	wg.Wait()
	c.reset()
	return nil
}

// retrieve the current sync index for this user from the server
func (c *Client) GetServerIdx(gen bool) (*svc.SyncIndex, error) {
	var endpoint string
	if gen {
		endpoint = c.Endpoints["gen index"]
	} else {
		endpoint = c.Endpoints["get index"]
	}
	resp, err := c.Client.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to contact server: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		c.dump(resp, true)
		return nil, fmt.Errorf("failed to get server sync index: %v", resp.StatusCode)
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return nil, err
	}
	var idx = new(svc.SyncIndex)
	if err = json.Unmarshal(buf.Bytes(), &idx); err != nil {
		return nil, err
	}
	return idx, nil
}

// ------- single-operation pushes and pulls from the server -------------

// send a known file to the server. For new files, use PushNewFile() instead.
func (c *Client) PushFile(file *svc.File) error {
	if err := c.Transfer.Upload(
		http.MethodPut, file, file.Endpoint,
	); err != nil {
		return err
	}
	return nil
}

// send a new file to the server. for updats to existing files,
// use PushFile() instead.
func (c *Client) PushNewFile(file *svc.File) error {
	if err := c.Transfer.Upload(
		http.MethodPost, file, c.Endpoints["new file"],
	); err != nil {
		return err
	}
	return nil
}

// download a file from the server. this assumes the file is already on the server,
// and that the client is intendending to update the local version of this file.
//
// not intended for new files discovered on the server -- this will be handled by a
// separate function PullNewFiles()
func (c *Client) PullFile(file *svc.File) error {
	if err := c.Transfer.Download(file.ClientPath, file.Endpoint); err != nil {
		return err
	}
	return nil
}
