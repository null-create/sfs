package client

import (
	"fmt"
	"time"

	"github.com/sfs/pkg/monitor"
)

// ---- file monitoring operations

// add a file listener to the map if the file isn't already present.
// will be a no-op if its already being watched.
func (c *Client) WatchFile(filePath string) {
	if !c.Monitor.Exists(filePath) {
		c.Monitor.WatchFile(filePath)
	}
}

// start monitoring files for changes
func (c *Client) StartMonitor() error {
	if err := c.Monitor.Start(c.Drive.DriveRoot); err != nil {
		return err
	}
	return nil
}

// stop all event listeners for this client
func (c *Client) StopMonitoring() error {
	if err := c.Monitor.ShutDown(); err != nil {
		return err
	}
	return nil
}

// add a new event handler for the given file
func (c *Client) NewHandler(fileID string) error {
	if _, exists := c.Handlers[fileID]; !exists {
		c.Handlers[fileID] = EventHandler
	} else {
		return fmt.Errorf("file (id=%v) is already registered", fileID)
	}
	return nil
}

// start an event handler for a given file
func (c *Client) StartHandler(fileID string) error {
	if handler, exists := c.Handlers[fileID]; exists {
		if err := handler(c, fileID); err != nil {
			return err
		}
	}
	return nil
}

// build a map of event handlers for client files.
// each handler will listen for events from files and will
// call synchronization operations accordingly
//
// should ideally only be called once during initialization
func (c *Client) BuildHandlers() error {
	// get list of files for the user
	files := c.Drive.GetFiles()

	// build handlers for each, populate handler map
	for _, file := range files {
		if _, exists := c.Handlers[file.ID]; !exists {
			c.Handlers[file.ID] = EventHandler
		}
	}
	return nil
}

// sets up a new listener for a given path.
//
// handles received events and starts transfer operations
func EventHandler(c *Client, filePath string) error {
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
		evts := monitor.NewEvents(false)
		for {
			select {
			case e := <-evt:
				switch e.Type {
				case monitor.FileCreate:
					c.Drive.SyncIndex.LastSync[fileID] = time.Now().UTC()
				case monitor.FileChange:
					c.Drive.SyncIndex.LastSync[fileID] = time.Now().UTC()
				case monitor.FileDelete:
					off <- true // shutdown monitoring thread
					delete(c.Drive.SyncIndex.LastSync, fileID)
				}
				evts.AddEvent(e)
				if evts.StartSync {
					// populate ToUpdate map before transferring all files to be synced
					// to the server
					c.Drive.SyncIndex = c.Drive.Root.WalkU(c.Drive.SyncIndex)

					evts.Reset()
				}
			default:
				continue
			}
		}
	}()
	return nil
}
