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

	"github.com/sfs/pkg/logger"
	svc "github.com/sfs/pkg/service"
)

// whether we should save to local storage, or push files to server.
func (c *Client) LocalSyncOnly() bool { return c.Conf.LocalBackup }

// resets client side sync mechanisms with a
// new baseline for item last sync times.
func (c *Client) reset() {
	c.Drive.SyncIndex.Reset()
	c.BuildSyncIndex()
}

// display server response clearly
func (c *Client) dump(resp *http.Response, body bool) {
	b, err := httputil.DumpResponse(resp, body)
	if err != nil {
		c.log.Error("failed to dump http response: " + err.Error())
	} else {
		if resp.StatusCode != http.StatusOK {
			c.log.Warn("request failed: " + resp.Status + "\n" + string(b) + "\n")
		} else {
			c.log.Log(logger.INFO, "server response: "+resp.Status)
		}
	}
}

// build a new client sync index. re-initializes
// client roots index if called.
func (c *Client) BuildSyncIndex() {
	// initialize sync index
	c.Drive.SyncIndex = svc.NewSyncIndex(c.UserID)

	// get any files
	files := c.Drive.GetFiles()
	if len(files) == 0 {
		c.log.Log(logger.WARN, "no files. nothing to index.")
		return
	}

	// NOTE: for future implementations when we can support monitoring directories
	//
	// get any directories
	// dirs, err := c.Db.GetUsersDirectories(c.UserID)
	// if err != nil {
	// 	c.log.Error(fmt.Sprintf("failed to get users directories: %s", err))
	// 	return
	// }
	// if len(dirs) == 0 {
	// 	c.log.Warn("no directories. nothing to index")
	// }

	// NOTE: the dir arg is set to nil until dir monitoring is supported
	c.Drive.SyncIndex = svc.BuildSyncIndex(files, nil, c.Drive.SyncIndex)
	c.log.Log(logger.INFO, fmt.Sprintf("%d files have been indexed", len(files)))
}

type SyncItems struct {
	pull []*svc.File
	push []*svc.File
}

// sync items between the client and the server.
//
// NOTE: this assumes that both the client and the server have
// a record of the objects to sync. if the server has a file the client doesn't
// know about, then this doesn't handle it, and vice-versa
func (c *Client) ServerSync() error {
	svrIdx, err := c.GetServerIdx(true) // get latest server sync index
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

	// pull latest versions of files from the server
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

	// push latest versions of files to the server
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
// server, then upload each in their own goroutines one batch at a time.
// each file is assumed to be already registered with the server, otherwise
// this will receive a 404 response and the upload will fail.
func (c *Client) Push() error {
	if len(c.Drive.SyncIndex.FilesToUpdate) == 0 {
		c.log.Warn("no files marked for uploading. sync index update map is empty")
		return nil
	}
	q := svc.BuildQ(c.Drive.SyncIndex)
	if q == nil {
		return fmt.Errorf("unable to build queue: no files found for syncing")
	}
	var wg sync.WaitGroup
	for len(q.Queue) > 0 {
		batch := q.Dequeue()
		for _, file := range batch.Files {
			wg.Add(1)
			go func() {
				defer wg.Done()
				c.log.Info(fmt.Sprintf("uploading %s...", file.Name))
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
func (c *Client) Pull() error {
	if len(c.Drive.SyncIndex.FilesToUpdate) == 0 {
		c.log.Warn("sync index update map has no contents. nothing to pull")
		return nil
	}
	q := svc.BuildQ(c.Drive.SyncIndex)
	if q == nil {
		return fmt.Errorf("unable to build queue: no files found for syncing")
	}
	var wg sync.WaitGroup
	for len(q.Queue) > 0 {
		batch := q.Dequeue()
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
	if err := c.Transfer.Upload(http.MethodPut, file, file.Endpoint); err != nil {
		return err
	}
	return nil
}

// send a new file to the server. for updats to existing files,
// use PushFile() instead.
func (c *Client) PushNewFile(file *svc.File) error {
	if err := c.Transfer.Upload(http.MethodPost, file, c.Endpoints["new file"]); err != nil {
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

// ---------------- local synchronization operations ----------------

// build sync index of all locally monitored files, then
// create backups in the backup directory of everything.
func (c *Client) LocalSync() error {
	files, err := c.Db.GetUsersFiles(c.UserID)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return nil
	}
	var wg sync.WaitGroup
	c.Drive.SyncIndex = svc.BuildToUpdate(files, nil, c.Drive.SyncIndex)
	for _, file := range c.Drive.SyncIndex.FilesToUpdate {
		wg.Add(1)
		go func() {
			if err := c.BackupFile(file); err != nil {
				c.log.Error(err.Error())
			}
			wg.Done()
		}()
	}
	wg.Wait()
	c.reset()
	return nil
}

func (c *Client) BackupFile(file *svc.File) error {
	if err := file.Copy(file.BackupPath); err != nil {
		return err
	}
	return nil
}

func (c *Client) BackupDir(dir *svc.Directory) error {
	if len(dir.Files) == 0 {
		c.log.Log(logger.INFO, fmt.Sprintf("directory '%s' has no files to backup", dir.Name))
	}
	if len(dir.Dirs) == 0 {
		c.log.Info(fmt.Sprintf("dir '%s' has no subdirectories to backup", dir.Name))
		return nil
	}
	return dir.CopyDir(dir.GetPath(), dir.BackupPath)
}
