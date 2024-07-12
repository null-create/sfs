package client

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sfs/pkg/logger"
	"github.com/sfs/pkg/monitor"
	svc "github.com/sfs/pkg/service"
)

// ---- file monitoring operations

// start monitoring files for changes.
//
// NOTE: we don't monitor files under sfs root because
// that is a desigated backup directory.
func (c *Client) StartMonitor() error {
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
	// NOTE: monitoring directories is currently not supported.
	// leaving this for future implementation iterations.
	// dirs := c.Drive.GetDirs()
	// dirs, err := c.Db.GetUsersDirectories(c.UserID)
	// if err != nil {
	// 	return err
	// }
	// for _, d := range dirs {
	// 	if err := c.WatchItem(d.Path); err != nil {
	// 		return err
	// 	}
	// }
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
// a dedicated event handler for the new file monitoring goroutine.
func (c *Client) WatchItem(path string) error {
	isDir, err := c.Monitor.IsDir(path)
	if err != nil {
		return err
	}
	// directory monitoring is not supported for the moment.
	if isDir {
		return fmt.Errorf(
			"%s is a directory. directory monitoring is not supporetd",
			filepath.Dir(path),
		)
	}
	if err := c.Monitor.Watch(path); err != nil {
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
func (c *Client) setupHandler(itemPath string) (chan monitor.Event, *monitor.Events, string, error) {
	evtChan := c.Monitor.GetEventChan(itemPath)

	thing, err := os.Stat(itemPath)
	if err != nil {
		return nil, nil, "", err
	}
	var id string
	if thing.IsDir() {
		itemID, err := c.Db.GetDirIDFromPath(itemPath)
		if err != nil {
			return nil, nil, "", err
		}
		if itemID == "" {
			return nil, nil, "", fmt.Errorf("no id found for directory %s", itemPath)
		}
		id = itemID
	} else {
		itemID, err := c.Db.GetFileIDFromPath(itemPath)
		if err != nil {
			return nil, nil, "", err
		}
		if itemID == "" {
			return nil, nil, "", fmt.Errorf("no id found for file %s", itemPath)
		}
		id = itemID
	}
	if evtChan == nil || id == "" {
		return nil, nil, "", fmt.Errorf(
			"failed to get param: evt=%v id=%s",
			evtChan, id,
		)
	}
	// events buffer.
	// used for managing and triggering synchronization
	// events between the client and the server.
	evts := monitor.NewEvents(cfgs.BufferedEvents)

	return evtChan, evts, id, nil
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
	c.log.Info(fmt.Sprintf("starting %d file handler(s)...", len(files)))
	for _, f := range files {
		if err := c.StartHandler(f.Path); err != nil {
			return err
		}
	}
	// NOTE: directory monitoring is not supported, but may be
	// in future versions, so this code will stay here for now.
	// start directory handlers
	// dirs := c.Drive.GetDirs()
	// dirs, err := c.Db.GetUsersDirectories(c.UserID)
	// if err != nil {
	// 	return nil
	// }
	// c.log.Info(fmt.Sprintf("starting %d directory handler(s)...", len(dirs)))
	// for _, d := range dirs {
	// 	if err := c.StartHandler(d.Path); err != nil {
	// 		return err
	// 	}
	// }
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
	c.log.Info(fmt.Sprintf("shutting down %d event handlers...", len(c.OffSwitches)))
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
	// NOTE: monitoring directories is not supported, but may be in the future.
	// os.ReadDir() was using a lot of resources.
	// dirs, err := c.Db.GetUsersDirectories(c.UserID)
	// if err != nil {
	// 	return err
	// }
	// for _, dir := range dirs {
	// 	if _, exists := c.Handlers[dir.Path]; !exists {
	// 		if err := c.NewEHandler(dir.Path); err != nil {
	// 			return err
	// 		}
	// 	}
	// }
	// // watch root
	// if err := c.WatchItem(c.Drive.Root.Path); err != nil {
	// 	return err
	// }
	return nil
}

// build a new event handler for a given file. does not start the handler,
// only adds it (and its offswitch) to the handlers map.
func (c *Client) NewEHandler(path string) error {
	offSwitch := make(chan bool)
	handler := func() {
		go func() {
			if err := c.handler(path, offSwitch); err != nil {
				c.log.Error(fmt.Sprintf("handler for %s failed: %v", filepath.Base(path), err))
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
	evtChan, evtBuf, itemID, err := c.setupHandler(itemPath)
	if err != nil {
		return err
	}
	// main listening loop for events
	for {
		select {
		case <-stop:
			c.log.Log(logger.INFO, fmt.Sprintf("stopping handler for %s...", filepath.Base(itemPath)))
			return nil
		case evt := <-evtChan:
			switch evt.Type {
			// NOTE: this is no longer supported, but may be used in the future versions.
			//
			// new files or directories were added to a monitored directory
			// case monitor.Add:
			// for _, eitem := range evt.Items {
			// 	// if this is a known item and it has just changed locations,
			// 	// then we just need to update the metadata, otherwise
			// 	// create a new object and register
			// 	if c.KnownItem(eitem.Path()) {
			// 		if err := c.UpdateItem(eitem); err != nil {
			// 			c.log.Error(fmt.Sprintf("failed to update item %s: %v", eitem.Path(), err))
			// 		}
			// 	}
			// 	// otherwise create new object and register
			// 	if eitem.IsDir() {
			// 		newDir := svc.NewDirectory(eitem.Name(), c.UserID, c.DriveID, eitem.Path())
			// 		if err := c.AddDirWithID(itemID, newDir); err != nil {
			// 			c.log.Error(fmt.Sprintf("failed to add new directory: %v", err))
			// 		}
			// 	} else {
			// 		newFile := svc.NewFile(eitem.Name(), c.DriveID, c.UserID, eitem.Path())
			// 		if err := c.AddFileWithDirID(itemID, newFile); err != nil {
			// 			c.log.Error(fmt.Sprintf("failed to add new file: %v", err))
			// 		}
			// 	}
			// }
			// NOTE: these new items may not be pushed to the server
			// since they will be added with their initial last sync times.
			// they will be added to the server after some modifications are detected,
			// and if auto sync is enabled.
			// evtBuf.AddEvent(evt)
			// case monitor.Remove:
			// TODO: handle for cases when items are removed from a directory
			// and possibly moved to another location.
			//
			// item name change
			case monitor.Name:
				if err := c.apply(evt.Path, "name"); err != nil {
					c.log.Error(fmt.Sprintf("failed to apply action: %v", err))
					break
				}
				evtBuf.AddEvent(evt)
			// item mode change
			case monitor.Mode:
				if err := c.apply(evt.Path, "mode"); err != nil {
					c.log.Error(fmt.Sprintf("failed to apply action: %v", err))
					break
				}
				evtBuf.AddEvent(evt)
			// item size changed
			case monitor.Size:
				if err := c.apply(evt.Path, "size"); err != nil {
					c.log.Error(fmt.Sprintf("failed to apply action: %v", err))
					break
				}
				evtBuf.AddEvent(evt)
			// item mod time change
			case monitor.ModTime:
				if err := c.apply(evt.Path, "modtime"); err != nil {
					c.log.Error(fmt.Sprintf("failed to apply action: %v", err))
					break
				}
				evtBuf.AddEvent(evt)
			// items content change
			case monitor.Change:
				if err := c.apply(evt.Path, "change"); err != nil {
					c.log.Error(fmt.Sprintf("failed to apply action: %v", err))
					break
				}
				evtBuf.AddEvent(evt)
			case monitor.Delete:
				c.log.Log(logger.INFO, fmt.Sprintf("handler for item (id=%s) stopping. item was deleted", itemID))
				return nil
			case monitor.Error:
				c.log.Warn(fmt.Sprintf("monitor for item (id=%s) encountered an error. stopping handler", itemID))
				return nil
			}
			// *** trigger synchronization operations once the event buffer has reached capacity ***
			if evtBuf.AtCap {
				// build update map and push changes if auto sync is enabled.
				c.Drive.SyncIndex = svc.BuildRootToUpdate(c.Drive.Root, c.Drive.SyncIndex)
				if c.autoSync() {
					/*
						TODO:

						we may need to make the client side auto back up system
						default to a *local* file system, rather than default to
						saving via the local network.

						Need to research:

							connecting two local computers via a shared
								wifi connection
								LAN connection

						If we want to continue prioritizing saving files via an https API,
						rather than just crud with the local system iteself. This solution
						may be a bit over-engineered, at the moment. It could also be a lot
						simpler.

						Maybe the client to server implementation could just be a mode setting,
						and we can default to saving everything under sfs/pkg/run/<user-name>
						while recreating directory structures.
					*/
					if !c.localBackup() {
						if err := c.Push(); err != nil {
							return err
						}
					} else {
						// local backup operations
						if c.IsDir(evt.Path) {
							d, err := c.GetDirByPath(evt.Path) // event paths are client side
							if err != nil {
								return err
							}
							if err := c.BackupDir(d); err != nil {
								return err
							}
						} else {
							f, err := c.GetFileByPath(evt.Path) // TODO: check which path this is (server or client)
							if err != nil {
								return err
							}
							if err := c.BackupFile(f); err != nil {
								return err
							}
						}
					}
				}
				evtBuf.Reset()                  // reset events buffer
				time.Sleep(monitor.WAIT_LONGER) // wait before resuming event handler
			}
		default:
			time.Sleep(monitor.WAIT)
			continue
		}
	}
}

// apply the given action
func (c *Client) apply(itemPath string, action string) error {
	item, err := os.Stat(itemPath)
	if err != nil {
		return err
	}
	if item.IsDir() {
		// *** NOTE: Directory actions are not supported
		// dir, err := c.GetDirByPath(itemPath)
		// if err != nil {
		// 	return err
		// }
		// switch action {
		// case "name":
		// 	dir.Name = item.Name()
		// 	if err := c.UpdateDirectory(dir); err != nil {
		// 		return err
		// 	}
		// case "size":
		// 	dir.Size = item.Size()
		// 	if err := c.UpdateDirectory(dir); err != nil {
		// 		return err
		// 	}
		// case "modtime":
		// 	dir.LastSync = item.ModTime()
		// 	if err := c.UpdateDirectory(dir); err != nil {
		// 		return err
		// 	}
		// case "change":
		// 	break
		// case "delete":
		// 	if err := c.RemoveDir(dir); err != nil {
		// 		return err
		// 	}
		// default:
		// 	return nil
		// }
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
