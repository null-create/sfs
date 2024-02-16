package client

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/sfs/pkg/monitor"
	svc "github.com/sfs/pkg/service"
)

// ---- file monitoring operations

// start monitoring files for changes
func (c *Client) StartMonitor() error {
	if err := c.Monitor.Start(c.Drive.DriveRoot); err != nil {
		return err
	}
	return nil
}

// stop all event listeners for this client.
// will be a no-op if there's no active monitoring threads.
func (c *Client) StopMonitoring() {
	c.Monitor.ShutDown()
}

// adds a file to monitor, then creates and starts
// a dedicated event listener and handler for this file.
func (c *Client) WatchItem(path string) error {
	if err := c.Monitor.WatchItem(path); err != nil {
		return err
	}
	if err := c.NewHandler(path); err != nil {
		return err
	}
	if err := c.StartHandler(path); err != nil {
		return err
	}
	return nil
}

// add a new event handler for the given file.
// path to the given file must already have a monitoring
// goroutine in place (call client.WatchFile(filePath) first).
func (c *Client) NewHandler(path string) error {
	if _, exists := c.Handlers[path]; !exists {
		if err := c.NewEHandler(path); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("%s is already registered", filepath.Base(path))
	}
	return nil
}

// get alll the necessary things for the event handler to operate independently
func (c *Client) setupHandler(itemPath string) (chan monitor.Event, chan bool, string, *monitor.Events, error) {
	evtChan := c.Monitor.GetEventChan(itemPath)
	offSwitch := c.Monitor.GetOffSwitch(itemPath)
	itemID, err := c.Db.GetFileID(itemPath) // NOTE this is only for files! not directories
	if err != nil {
		return nil, nil, "", nil, err
	}
	if evtChan == nil || offSwitch == nil || itemID == "" {
		return nil, nil, "", nil, fmt.Errorf(
			"failed to get param: evt=%v off=%v id=%s",
			evtChan, offSwitch, itemID,
		)
	}
	evts := monitor.NewEvents(cfgs.BufferedEvents)
	return evtChan, offSwitch, itemID, evts, nil
}

// start an event handler for a given file.
// will be a no-op if the handler does not exist.
func (c *Client) StartHandler(path string) error {
	if handler, exists := c.Handlers[path]; exists {
		handler()
	}
	return nil
}

// start all available handlers
func (c *Client) StartHandlers() error {
	files := c.Drive.GetFiles()
	if len(files) == 0 {
		log.Print("[INFO] no files to start handlers for")
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
func (c *Client) BuildHandlers() error {
	files := c.Drive.GetFiles()
	if len(files) == 0 {
		log.Print("[INFO] no files to build handlers for")
		return nil
	}
	for _, file := range files {
		if _, exists := c.Handlers[file.Path]; !exists {
			if err := c.NewEHandler(file.Path); err != nil {
				return err
			}
		}
	}
	return nil
}

// build a new event handler for a given file. does not start the handler,
// only adds it (and its offswitch) to the handlers map.
func (c *Client) NewEHandler(path string) error {
	// handler off-switch
	offSwitch := make(chan bool)
	// handler
	handler := func() {
		// start listener
		go func() {
			if err := c.listener(path, offSwitch); err != nil {
				log.Printf("[ERROR] listener failed: %v", err)
				// shut down monitoring thread for this event handler.
				// all monitoring threads must have a dedicated handler.
				c.Monitor.CloseChan(path)
			}
		}()
	}
	c.Handlers[path] = handler
	c.OffSwitches[path] = offSwitch
	return nil
}

// dedicated listener for item events.
// items can be either files or directories.
func (c *Client) listener(path string, stop chan bool) error {
	// get all necessary params for the handler
	evtChan, off, itemID, evts, err := c.setupHandler(path)
	if err != nil {
		return err
	}
	// main listening loop for events
	for {
		select {
		case <-stop:
			log.Printf("[INFO] stopping event handler for item id=%v ...", itemID)
			return nil
		case e := <-evtChan:
			switch e.Type {
			// new files or directories were added to a monitored directory
			case monitor.Add:
				for _, eitem := range e.Items {
					item, err := os.Stat(eitem.Path)
					if err != nil {
						log.Printf("[ERROR] failed to get item information: %v", err)
					}
					if item.IsDir() {
						newDir := svc.NewDirectory(eitem.Name(), c.UserID, c.DriveID, eitem.Path)
						if err := c.AddDir(newDir.ID, newDir); err != nil {
							log.Printf("[ERROR] failed to add new directory: %v", err)
						}
					} else {
						newFile := svc.NewFile(eitem.Name(), c.DriveID, c.UserID, e.Path)
						if err := c.AddFile(itemID, newFile); err != nil {
							log.Printf("[ERROR] failed to add new file: %v", err)
						}
					}
				}
				evts.AddEvent(e)
			// item name change
			case monitor.Name:
				c.applyChange(e.Path, "name")
				evts.AddEvent(e)
			// item mode change
			case monitor.Mode:
				c.applyChange(e.Path, "mode")
				evts.AddEvent(e)
			// item size changevedbooboo
			case monitor.Size:
				c.applyChange(e.Path, "size")
				evts.AddEvent(e)
			// item mod time change
			case monitor.ModTime:
				c.applyChange(e.Path, "modtime")
				evts.AddEvent(e)
			// items content change
			case monitor.Change:
				c.applyChange(e.Path, "change")
				evts.AddEvent(e)
			case monitor.Delete:
				off <- true // shutdown monitoring thread, remove from index, and shut down handler
				log.Printf("[INFO] handler for item (id=%s) stopping. item was deleted.", itemID)
				return nil
			}
			// TODO: need to decide how ofter to run sync operations once the
			// events buffer reaches capacity (i ->n).
			// should have some configs around whether we build the update map
			// every time, or if its a single event. BuildToUpdate is mainly intended
			// for large sync operations with files and directories processsed in batches.
			//
			// NOTE: whatever operations take place here will need to be thread safe!
			if evts.AtCap {
				// build update map and push file changes to server
				c.Drive.SyncIndex = svc.BuildToUpdate(c.Drive.Root, c.Drive.SyncIndex)
				if err := c.Push(); err != nil {
					return err
				}
				evts.Reset()             // resets events buffer
				time.Sleep(monitor.WAIT) // wait before resuming event handler
			}
		default:
			continue
		}
	}
}

// applyChange the given action using the supplied event object
func (c *Client) applyChange(itemPath string, action string) error {
	item, err := os.Stat(itemPath)
	if err != nil {
		return err
	}
	if item.IsDir() {
		dir, err := c.GetDirByPath(itemPath)
		if err != nil {
			return err
		}
		switch action {
		case "name":
			dir.Name = item.Name()
			if err := c.UpdateDirectory(dir); err != nil {
				return err
			}
		case "size":
			dir.Size = item.Size()
			if err := c.UpdateDirectory(dir); err != nil {
				return err
			}
		case "modtime":
			dir.LastSync = item.ModTime()
			if err := c.UpdateDirectory(dir); err != nil {
				return err
			}
		case "change":
		case "delete":
			if err := c.RemoveDir(dir.ID); err != nil {
				return err
			}
		default:
			return nil
		}
	} else {
		file, err := c.GetFileByPath(itemPath)
		if err != nil {
			return err
		}
		switch action {
		case "name":
			file.Name = item.Name()
			if err := c.UpdateFile(file); err != nil {
				return err
			}
		case "mode":
			file.Mode = item.Mode()
			if err := c.UpdateFile(file); err != nil {
				return err
			}
		case "size":
			file.Size = item.Size()
			if err := c.UpdateFile(file); err != nil {
				return err
			}
		case "modtime":
			file.LastSync = item.ModTime()
			if err := c.UpdateFile(file); err != nil {
				return err
			}
		case "change":
		case "delete":
			if err := c.RemoveFile(file.DirID, file); err != nil {
				return err
			}
		default:
			return nil
		}
	}
	return nil
}
