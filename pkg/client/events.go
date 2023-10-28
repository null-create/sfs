package client

import (
	"fmt"
	"time"

	"github.com/sfs/pkg/monitor"
)

// ---- file monitoring operations

// populates c.Monitor.Events with listener goroutines for all
// files in the client's drive.
func (c *Client) WatchFiles() error {
	return c.Monitor.WatchAll(c.Drive.Root.Path)
}

// add a file listener to the map if the file isn't already present.
// will be a no-op if its already being watched.
func (c *Client) WatchFile(filePath string) {
	if _, exists := c.Monitor.Events[filePath]; !exists {
		c.Monitor.WatchFile(filePath)
	}
}

// stop all event listeners for this client
func (c *Client) StopMonitoring() error {
	if err := c.Monitor.ShutDown(); err != nil {
		return err
	}
	return nil
}

// main event loop that coordinates sync operations after
// receiving a file event from the listener.

// TODO: this should be part of a larger data structure
// that is coordinating (or at least keeping track of) all
// the watcher/listener event goroutines
func (c *Client) EventHandler(filePath string) error {
	// get event channel for this file to listen to
	evt := c.Monitor.GetEventChan(filePath)
	if evt == nil {
		return fmt.Errorf("no event listener for file %s", filePath)
	}
	// get off switch for this monitor
	off := c.Monitor.GetOffSwitch(filePath)
	if off == nil {
		return fmt.Errorf("no shut off channel for file %s", filePath)
	}
	// get ID for this file so we can quickly update the sync index
	// without having to rebuild it every time
	fileID, err := c.Db.GetFileID(filePath)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case e := <-evt:
				switch e.Type {
				// TODO: add sync operations
				case monitor.FileCreate:
					c.Drive.SyncIndex.LastSync[fileID] = time.Now().UTC()
				case monitor.FileChange:
					c.Drive.SyncIndex.LastSync[fileID] = time.Now().UTC()
				case monitor.FileDelete:
					off <- true // shutdown monitoring thread
					delete(c.Drive.SyncIndex.LastSync, fileID)
				}
			default:
				continue
			}
		}
	}()

	return nil
}
