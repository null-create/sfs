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
	// monitor files under sfs root
	if err := c.Monitor.Start(c.Root); err != nil {
		return err
	}
	// monitor all other registered items distributed
	// in the users system
	// files := c.Drive.GetFiles()
	files, err := c.Db.GetUsersFiles(c.UserID)
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := c.WatchItem(f.ClientPath); err != nil {
			return err
		}
	}
	// dirs := c.Drive.GetDirs()
	dirs, err := c.Db.GetUsersDirectories(c.UserID)
	if err != nil {
		return err
	}
	for _, d := range dirs {
		if err := c.WatchItem(d.Path); err != nil {
			return err
		}
	}
	return nil
}

// stop all event monitors for this client.
// will be a no-op if there's no active monitoring threads.
func (c *Client) StopMonitoring() {
	c.Monitor.ShutDown()
}

// initialize handlers and monitor off switch maps
func (c *Client) InitHandlerMaps() {
	c.Handlers = make(map[string]func())
	c.OffSwitches = make(map[string]chan bool)
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

	thing, err := os.Stat(itemPath)
	if err != nil {
		return nil, nil, "", nil, err
	}
	var id string
	if thing.IsDir() {
		itemID, err := c.Db.GetDirIDFromPath(itemPath)
		if err != nil {
			return nil, nil, "", nil, err
		}
		if itemID == "" {
			return nil, nil, "", nil, fmt.Errorf("no id found for directory %s", itemPath)
		}
		id = itemID
	} else {
		itemID, err := c.Db.GetFileIDFromPath(itemPath)
		if err != nil {
			return nil, nil, "", nil, err
		}
		if itemID == "" {
			return nil, nil, "", nil, fmt.Errorf("no id found for file %s", itemPath)
		}
		id = itemID
	}
	if evtChan == nil || offSwitch == nil || id == "" {
		return nil, nil, "", nil, fmt.Errorf(
			"failed to get param: evt=%v off=%v id=%s",
			evtChan, offSwitch, id,
		)
	}
	// events buffer.
	// used for managing and triggering synchronization
	// events between the client and the server.
	evts := monitor.NewEvents(cfgs.BufferedEvents)

	return evtChan, offSwitch, id, evts, nil
}

// start an event handler for a given file.
// will be a no-op if the handler does not exist.
func (c *Client) StartHandler(path string) error {
	if handler, exists := c.Handlers[path]; exists {
		handler()
	}
	return nil
}

// start all available listeners
func (c *Client) StartHandlers() error {
	// start file handlers
	// files := c.Drive.GetFiles()
	files, err := c.Db.GetUsersFiles(c.UserID)
	if err != nil {
		return err
	}
	log.Printf("[INFO] starting %d file handler(s)...", len(files))
	for _, f := range files {
		if err := c.StartHandler(f.Path); err != nil {
			return err
		}
	}
	// start directory handlers
	// dirs := c.Drive.GetDirs()
	dirs, err := c.Db.GetUsersDirectories(c.UserID)
	if err != nil {
		return nil
	}
	log.Printf("[INFO] starting %d directory handler(s)...", len(dirs))
	for _, d := range dirs {
		if err := c.StartHandler(d.Path); err != nil {
			return err
		}
	}
	return nil
}

// stop an existening event listener.
func (c *Client) StopHandler(itemPath string) error {
	if _, ok := c.OffSwitches[itemPath]; ok {
		c.OffSwitches[itemPath] <- true
	}
	if _, ok := c.Handlers[itemPath]; ok {
		c.Handlers[itemPath] = nil
	}
	return nil
}

// stops all event handler goroutines.
func (c *Client) StopHandlers() {
	log.Printf("[INFO] shutting down %d event handlers...", len(c.OffSwitches))
	for path := range c.Handlers {
		c.Handlers[path] = nil
	}
}

// build a map of event handlers for client files and directories.
// each handler will listen for events from the users items and will
// call synchronization operations accordingly, assuming auto sync is enabled.
// if no files or directories are present during the build call then
// this will be a no-op.
//
// should ideally only be called once during initialization.
func (c *Client) BuildHandlers() error {
	// TODO: we should ideally be using c.Drive.GetFiles()
	files, err := c.Db.GetUsersFiles(c.UserID)
	if err != nil {
		return err
	}
	for _, file := range files {
		if _, exists := c.Handlers[file.Path]; !exists {
			if err := c.NewEHandler(file.Path); err != nil {
				return err
			}
		}
	}
	// TODO: we should ideally be using c.Drive.GetDirectories()
	dirs, err := c.Db.GetUsersDirectories(c.UserID)
	if err != nil {
		return err
	}
	for _, dir := range dirs {
		if _, exists := c.Handlers[dir.Path]; !exists {
			if err := c.NewEHandler(dir.Path); err != nil {
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
		go func() {
			if err := c.handler(path, offSwitch); err != nil {
				log.Printf("[ERROR] handler for %s failed: %v", filepath.Base(path), err)
			}
		}()
	}
	c.Handlers[path] = handler
	c.OffSwitches[path] = offSwitch
	return nil
}

// dedicated handler for item events.
// items can be either files or directories.
func (c *Client) handler(itemPath string, stop chan bool) error {
	// get all necessary params for the event handler.
	evtChan, off, itemID, evts, err := c.setupHandler(itemPath)
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
					item, err := os.Stat(eitem.Path())
					if err != nil {
						log.Printf("[ERROR] failed to get item information: %v", err)
					}
					if item.IsDir() {
						newDir := svc.NewDirectory(eitem.Name(), c.UserID, c.DriveID, eitem.Path())
						if err := c.AddDirWithID(newDir.ID, newDir); err != nil {
							log.Printf("[ERROR] failed to add new directory: %v", err)
						}
					} else {
						newFile := svc.NewFile(eitem.Name(), c.DriveID, c.UserID, e.Path)
						if err := c.AddFileWithID(itemID, newFile); err != nil {
							log.Printf("[ERROR] failed to add new file: %v", err)
						}
					}
				}
				// NOTE: these new items may not be pushed to the server
				// since they will be added with their initial last sync times.
				// they will be added to the server after some modifications are detected,
				// and if auto sync is enabled.
				evts.AddEvent(e)
			// item name change
			case monitor.Name:
				if err := c.apply(e.Path, "name"); err != nil {
					log.Printf("[ERROR] failed to apply action: %v", err)
					break
				}
				evts.AddEvent(e)
			// item mode change
			case monitor.Mode:
				if err := c.apply(e.Path, "mode"); err != nil {
					log.Printf("[ERROR] failed to apply action: %v", err)
					break
				}
				evts.AddEvent(e)
			// item size changed
			case monitor.Size:
				if err := c.apply(e.Path, "size"); err != nil {
					log.Printf("[ERROR] failed to apply action: %v", err)
					break
				}
				evts.AddEvent(e)
			// item mod time change
			case monitor.ModTime:
				if err := c.apply(e.Path, "modtime"); err != nil {
					log.Printf("[ERROR] failed to apply action: %v", err)
					break
				}
				evts.AddEvent(e)
			// items content change
			case monitor.Change:
				if err := c.apply(e.Path, "change"); err != nil {
					log.Printf("[ERROR] failed to apply action: %v", err)
					break
				}
				evts.AddEvent(e)
			case monitor.Delete:
				off <- true // shutdown monitoring thread, remove from index, and shut down handler
				log.Printf("[INFO] handler for item (id=%s) stopping. item was deleted.", itemID)
				return nil
			}
			// trigger synchronization operations once the event buffer has reached capacity
			// NOTE: whatever operations take place here will need to be thread safe!
			if evts.AtCap {
				// build update map and push changes if auto sync is enabled.
				c.Drive.SyncIndex = svc.BuildToUpdate(c.Drive.Root, c.Drive.SyncIndex)
				if c.autoSync() {
					if err := c.Push(); err != nil {
						return err
					}
				}
				evts.Reset()             // reset events buffer
				time.Sleep(monitor.WAIT) // wait before resuming event handler
			}
		default:
			continue
		}
	}
}

// apply the given action using the supplied event object
func (c *Client) apply(itemPath string, action string) error {
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
			break
		case "delete":
			if err := c.RemoveDir(dir); err != nil {
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
			break
		case "delete":
			if err := c.RemoveFile(file); err != nil {
				return err
			}
		default:
			return nil
		}
	}
	return nil
}
