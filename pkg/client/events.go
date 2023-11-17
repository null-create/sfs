package client

import (
	"fmt"
	"log"
	"time"

	"github.com/sfs/pkg/monitor"
)

type EHandler func(chan monitor.Event, chan bool, string, *monitor.Events) error

// ---- file monitoring operations

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

// add a file listener to the map if the file isn't already present.
// will be a no-op if its already being watched.
func (c *Client) WatchFile(filePath string) {
	if !c.Monitor.Exists(filePath) {
		c.Monitor.WatchFile(filePath)
	}
}

// add a new event handler for the given file
func (c *Client) NewHandler(filePath string) error {
	if _, exists := c.Handlers[filePath]; !exists {
		c.Handlers[filePath] = EventHandler
	} else {
		return fmt.Errorf("file (%v) is already registered", filePath)
	}
	return nil
}

// get alll the necessary things for the event handler to operate independently
func (c *Client) setupHandler(filePath string) (chan monitor.Event, chan bool, string, *monitor.Events, error) {
	evt := c.Monitor.GetEventChan(filePath)
	off := c.Monitor.GetOffSwitch(filePath)
	fileID, err := c.Db.GetFileID(filePath)
	if err != nil {
		return nil, nil, "", nil, err
	}
	if evt == nil || off == nil || fileID == "" {
		return nil, nil, "", nil, fmt.Errorf("failed to get param: %v %v %s", evt, off, fileID)
	}
	// TODO: buffered events should be a client setting
	evts := monitor.NewEvents(false)
	return evt, off, fileID, evts, nil
}

// start an event handler for a given file
func (c *Client) StartHandler(filePath string) error {
	if handler, exists := c.Handlers[filePath]; exists {
		evt, off, fileID, evts, err := c.setupHandler(filePath)
		if err != nil {
			return err
		}
		if err := handler(evt, off, fileID, evts); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) EventInfo(evt monitor.Event) string {
	msg := fmt.Sprintf(
		"[INFO] file event -> time: %v | type: %s | path: %s",
		evt.ID, evt.Type, evt.Path,
	)
	log.Print(msg)
	return msg
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

// handles received events and starts transfer operations
func EventHandler(evt chan monitor.Event, off chan bool, fileID string, evts *monitor.Events) error {
	go func() {
		for {
			select {
			case e := <-evt:
				// c.EventInfo(e) // display event info
				switch e.Type {
				case monitor.FileCreate:
					evts.AddEvent(e)
				case monitor.FileChange:
					evts.AddEvent(e)
				case monitor.FileDelete:
					off <- true // shutdown monitoring thread, remove from index, and shut down handler
					log.Printf("[INFO] handler for file (id=%s) stopping. file was deleted.", fileID)
					return
				}
				if evts.StartSync {
					log.Printf("[INFO] sync operation started at: %v", time.Now().UTC())
					// populate ToUpdate map only if *evts is buffered* before transferring all
					// files to be synced to the server, otherwise just sync the single file
					// (no need to update the entire ToUpdate map for one file)
					if evts.Buffered {
						// c.Drive.SyncIndex = c.Drive.Root.WalkU(c.Drive.SyncIndex)
					}
					// TODO: sync ops...
					//
					// will need to be able to use c.Drive.SyncIndex as part of its signature
					// will
					evts.Reset()
				}
			default:
				continue
			}
		}
	}()
	return nil
}
