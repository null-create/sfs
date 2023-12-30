package client

import (
	"fmt"
	"log"
	"time"

	"github.com/sfs/pkg/monitor"
)

type EHandler func(chan monitor.Event, chan bool, string, *monitor.Events) (chan bool, error)

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
	c.Monitor.WatchFile(filePath)
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
	evtChan := c.Monitor.GetEventChan(filePath)
	offSwitch := c.Monitor.GetOffSwitch(filePath)
	fileID, err := c.Db.GetFileID(filePath)
	if err != nil {
		return nil, nil, "", nil, err
	}
	if evtChan == nil || offSwitch == nil || fileID == "" {
		return nil, nil, "", nil, fmt.Errorf(
			"failed to get param: evt=%v off=%v fileID=%s",
			evtChan, offSwitch, fileID,
		)
	}
	// TODO: buffered events should be a client setting
	evts := monitor.NewEvents(false)
	return evtChan, offSwitch, fileID, evts, nil
}

// start an event handler for a given file
func (c *Client) StartHandler(filePath string) error {
	if handler, exists := c.Handlers[filePath]; exists {
		evt, off, fileID, evts, err := c.setupHandler(filePath)
		if err != nil {
			return err
		}
		if offSwitch, err := handler(evt, off, fileID, evts); err == nil {
			c.OffSwitches[filePath] = offSwitch
		} else {
			return err
		}
	}
	return nil
}

// start all available handlers
func (c *Client) StartHandlers() error {
	files := c.Drive.GetFiles()
	if len(files) == 0 {
		log.Print("[WARNING] no files to start handlers for")
		return nil
	}
	for _, f := range files {
		if err := c.StartHandler(f.Path); err != nil {
			return err
		}
	}
	return nil
}

// stops all event handler goroutines.
func (c *Client) StopHandlers() {
	log.Printf("[INFO] shutting down monitoring handlers...")
	for _, off := range c.OffSwitches {
		off <- true
	}
}

// build a map of event handlers for client files.
// each handler will listen for events from files and will
// call synchronization operations accordingly. if no files
// are present during the build call then this will be a no-op.
//
// should ideally only be called once during initialization
func (c *Client) BuildHandlers() {
	if c.Handlers == nil {
		c.Handlers = make(map[string]EHandler)
	}
	files := c.Drive.GetFiles()
	if len(files) == 0 {
		log.Print("[WARNING] no files to build handlers for")
		return
	}
	for _, file := range files {
		if _, exists := c.Handlers[file.Path]; !exists {
			c.Handlers[file.Path] = EventHandler
		}
	}
}

// handles received events from the client's monitor component
// and starts transfer operations by updating the sync doc flag.
func EventHandler(evt chan monitor.Event, off chan bool, fileID string, evts *monitor.Events) (chan bool, error) {
	stopMonitor := make(chan bool)
	go func() {
		for {
			select {
			case <-stopMonitor:
				log.Printf("[INFO] stopping event handler for file id=%v ...", fileID)
				return
			case e := <-evt:
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
					evts.Notify()            // sets sync doc to indicate sync operation
					evts.Reset()             // resets events buffer
					time.Sleep(monitor.WAIT) // wait before resuming event handler
				}
			default:
				continue
			}
		}
	}()
	return stopMonitor, nil
}
