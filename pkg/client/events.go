package client

import (
	"fmt"
	"log"

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
	evt := c.Monitor.GetEventChan(filePath)
	if evt == nil {
		return fmt.Errorf("no event listener for file %s", filePath)
	}
	off := c.Monitor.GetOffSwitch(filePath)
	if off == nil {
		return fmt.Errorf("no shut off channel for file %s", filePath)
	}
	go func() {
		for {
			select {
			case e := <-evt:
				switch e {
				// TODO: need a way to individually update
				// the sync indexes toUpdate map, rather than
				// recursively every time.
				case monitor.FileCreate:
					// c.Drive.SyncIndex.ToUpdate[filePath] = time.Now().UTC()
				case monitor.FileChange:

				case monitor.FileDelete:
				}
			case <-off:
				c.Monitor.OffSwitches[filePath] <- true // shut down monitor
				log.Printf("stopping event handler for file %s", filePath)
				return
			}
		}
	}()
	return nil
}
