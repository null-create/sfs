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

// stop all event monitors for this client.
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
	if err := c.NewListener(path); err != nil {
		return err
	}
	if err := c.StartListener(path); err != nil {
		return err
	}
	return nil
}

// add a new event handler for the given file.
// path to the given file must already have a monitoring
// goroutine in place (call client.WatchFile(filePath) first).
func (c *Client) NewListener(path string) error {
	if _, exists := c.Listeners[path]; !exists {
		if err := c.NewEListener(path); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("%s is already registered", filepath.Base(path))
	}
	return nil
}

// get alll the necessary things for the event handler to operate independently
func (c *Client) setupListener(itemPath string) (chan monitor.Event, chan bool, string, *monitor.Events, error) {
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
		itemID, err := c.Db.GetFileID(itemPath)
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

	// new events buffer.
	// used for triggering synchronization events between the
	// client and the server.
	evts := monitor.NewEvents(cfgs.BufferedEvents)

	return evtChan, offSwitch, id, evts, nil
}

// start an event handler for a given file.
// will be a no-op if the handler does not exist.
func (c *Client) StartListener(path string) error {
	if listener, exists := c.Listeners[path]; exists {
		listener()
	}
	return nil
}

// start all available listeners
func (c *Client) StartListeners() error {
	files := c.Drive.GetFiles()
	if len(files) == 0 {
		log.Print("[INFO] no files to start listeners for")
		return nil
	}
	for _, f := range files {
		if err := c.StartListener(f.Path); err != nil {
			return err
		}
	}
	return nil
}

// stop an existening event listener.
func (c *Client) StopListener(itemPath string) error {
	if _, ok := c.OffSwitches[itemPath]; ok {
		c.OffSwitches[itemPath] <- true
	}
	if _, ok := c.Listeners[itemPath]; ok {
		c.Listeners[itemPath] = nil
	}
	return nil
}

// stops all event handler goroutines.
func (c *Client) StopListeners() {
	log.Printf("[INFO] shutting down event listeners...")
	for _, off := range c.OffSwitches {
		off <- true
	}
	for path := range c.Listeners {
		c.Listeners[path] = nil
	}
	c.Wg.Wait()
}

// remove all listener instances
func (c *Client) DestroyListeners() {
	for path := range c.Listeners {
		c.Listeners[path] = nil
	}
}

// build a map of event handlers for client files.
// each handler will listen for events from files and will
// call synchronization operations accordingly. if no files
// are present during the build call then this will be a no-op.
//
// should ideally only be called once during initialization
func (c *Client) BuildListeners() error {
	// TODO: we should ideally be using c.Drive.GetFiles()
	files, err := c.Db.GetUsersFiles(c.UserID)
	if err != nil {
		return err
	}
	for _, file := range files {
		if _, exists := c.Listeners[file.Path]; !exists {
			if err := c.NewEListener(file.Path); err != nil {
				return err
			}
		}
	}
	return nil
}

// build a new event handler for a given file. does not start the handler,
// only adds it (and its offswitch) to the handlers map.
func (c *Client) NewEListener(path string) error {
	// listener off-switch
	offSwitch := make(chan bool)
	// listener
	listener := func() {
		// start listener
		go func() {
			if err := c.listener(path, offSwitch); err != nil {
				log.Printf("[ERROR] listener failed: %v", err)
			}
		}()
	}
	c.Wg.Add(1)
	c.Listeners[path] = listener
	c.OffSwitches[path] = offSwitch
	return nil
}

// dedicated listener for item events.
// items can be either files or directories.
func (c *Client) listener(itemPath string, stop chan bool) error {
	// get all necessary params for the event listener.
	evtChan, off, itemID, evts, err := c.setupListener(itemPath)
	if err != nil {
		return err
	}
	// main listening loop for events
	for {
		select {
		case <-stop:
			log.Printf("[INFO] stopping event handler for item id=%v ...", itemID)
			c.Wg.Done()
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
				// temporarily removed while were still testing...
				// if err := c.Push(); err != nil {
				// 	return err
				// }
				evts.Reset()             // resets events buffer
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
