package monitor

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/sfs/pkg/auth"
)

/*
this is the file for the background event-listener daemon.

this will listen for events like a file being saved within the client's drive, which will then
automatically start a new sync index operation. whether the user wants to automatically sync or not
should be a setting, but the daemon will automatically make a new sync index with each file or directory
modification.

should also have a mechanism to interrupt a sync operation if a new event occurs.
*/

// arbitrary wait time between checks
const WAIT = time.Millisecond * 500

type Monitor struct {
	// path to the users drive root to monitor
	Path string

	// map of channels to active listeners.
	// key is the absolute file path, value is the channel to the watchFile() thread
	// associated with that file
	//
	// key = file path, val is Event channel
	Events map[string]chan Event

	// map of channels to active listeners that will shut down the watcher goroutine
	// when set to true.
	//
	// key = file path, val is bool chan
	OffSwitches map[string]chan bool
}

func NewMonitor(drvRoot string) *Monitor {
	return &Monitor{
		Path:        drvRoot,
		Events:      make(map[string]chan Event),
		OffSwitches: make(map[string]chan bool),
	}
}

// see if an event channel exists for a given filepath.
func (m *Monitor) IsMonitored(path string) bool {
	if _, exists := m.Events[path]; exists {
		return true
	}
	return false
}

// recursively builds watchers for all files in the directory
// and subdirectories, then starts each event channel
func (m *Monitor) Start(rootpath string) error {
	if err := watchAll(rootpath, m); err != nil {
		return fmt.Errorf("failed to start monitor: %v", err)
	}
	return nil
}

// make sure the physical file or directory actually exists
func (m *Monitor) Exists(path string) bool {
	if _, err := os.Stat(path); err != nil && errors.Is(err, os.ErrNotExist) {
		return false
	} else if err != nil {
		log.Printf("[ERROR] failed to retrieve stat for: %s\n %v", path, err)
		return false
	}
	return true
}

// is this item a directory?
func (m *Monitor) IsDir(path string) (bool, error) {
	if stat, err := os.Stat(path); err == nil {
		return stat.IsDir(), nil
	} else if err != nil {
		return false, fmt.Errorf("failed to get stats for %v: %v", filepath.Base(path), err)
	}
	return false, nil
}

// add a file or directory to the events map and create a new monitoring
// goroutine. will need a corresponding events handler. will be a no-op if the
// given path is already being monitored.
func (m *Monitor) WatchItem(path string) error {
	// make sure this item actually exists
	if !m.Exists(path) {
		return fmt.Errorf("%s does not exist", filepath.Base(path))
	}
	if !m.IsMonitored(path) {
		isdir, err := m.IsDir(path)
		if err != nil {
			return err
		}
		// add stop channel before starting monitoring
		stop := make(chan bool)
		m.OffSwitches[path] = stop
		if isdir {
			m.Events[path] = watchDir(path, stop)
		} else {
			m.Events[path] = watchFile(path, stop)
		}
	}
	return nil
}

// get an event listener channel for a given file
func (m *Monitor) GetEventChan(filePath string) chan Event {
	if evtChan, exists := m.Events[filePath]; exists {
		return evtChan
	}
	log.Print("[ERROR] event channel not found")
	return nil
}

// get an off switch for a given monitoring goroutine.
// off switches, when set to true, will shut down the monitoring process.
func (m *Monitor) GetOffSwitch(filePath string) chan bool {
	if offSwitch, exists := m.OffSwitches[filePath]; exists {
		return offSwitch
	}
	log.Print("[ERROR] off switch not found for monitoring goroutine")
	return nil
}

// get a slice of file paths for associated monitoring goroutines.
// returns nil if none are available.
func (m *Monitor) GetPaths() []string {
	if len(m.Events) == 0 {
		return nil
	}
	paths := make([]string, 0, len(m.Events))
	for path := range m.Events {
		paths = append(paths, path)
	}
	return paths
}

// close a listener channel for a given file.
// will be a no-op if the file is not registered.
func (m *Monitor) CloseChan(filePath string) {
	if m.IsMonitored(filePath) {
		m.OffSwitches[filePath] <- true // shut down monitoring thread before closing
		delete(m.OffSwitches, filePath)
		delete(m.Events, filePath)
	}
}

// shutdown all active monitoring threads
func (m *Monitor) ShutDown() {
	if len(m.OffSwitches) == 0 {
		return
	}
	log.Printf("[INFO] shutting down %d active monitoring threads...", len(m.OffSwitches))
	for path := range m.OffSwitches {
		m.OffSwitches[path] <- true
	}
}

// ----------------------------------------------------------------------------

// creates a new monitor goroutine for a given file or directory.
// returns a channel that sends events to the listener for handling
func watchFile(filePath string, stop chan bool) chan Event {
	initialStat, err := os.Stat(filePath)
	if err != nil {
		log.Printf("[ERROR] failed to get initial info for %s: %v\nunable to monitor", filepath.Base(filePath), err)
		return nil
	}
	baseName := filepath.Base(filePath)

	// event channel used by the event handler goroutine
	evt := make(chan Event)

	// dedicated watcher function
	var watcher = func() {
		for {
			select {
			case <-stop:
				log.Printf("[INFO] shutting down monitoring for %s...", filePath)
				close(evt)
				return
			default:
				stat, err := os.Stat(filePath)
				if err != nil && err != os.ErrNotExist {
					log.Printf("[ERROR] %v\nstopping monitoring for %s...", err, baseName)
					close(evt)
					return
				}
				switch {
				// file deletion
				case err == os.ErrNotExist:
					log.Printf("[INFO] %s deleted. stopping monitoring...", baseName)
					evt <- Event{
						Type: Delete,
						ID:   auth.NewUUID(),
						Path: filePath,
					}
					close(evt)
					return
				// file size change
				case stat.Size() != initialStat.Size():
					log.Printf("[INFO] size change detected: %f kb -> %f kb", float64(initialStat.Size()/1000), float64(stat.Size()/1000))
					evt <- Event{
						Type: Size,
						ID:   auth.NewUUID(),
						Path: filePath,
					}
					initialStat = stat
				// file modification time change
				case stat.ModTime() != initialStat.ModTime():
					log.Printf("[INFO] mod time change detected: %v -> %v", initialStat.ModTime(), stat.ModTime())
					evt <- Event{
						Type: ModTime,
						ID:   auth.NewUUID(),
						Path: filePath,
					}
					initialStat = stat
				// file mode change
				case stat.Mode() != initialStat.Mode():
					log.Printf("[INFO] mode change detected: %v -> %v", initialStat.Mode(), stat.Mode())
					evt <- Event{
						Type: Mode,
						ID:   auth.NewUUID(),
						Path: filePath,
					}
					initialStat = stat
				// file name change
				case stat.Name() != initialStat.Name():
					log.Printf("[INFO] file name change detected: %v -> %v", initialStat.Name(), stat.Name())
					evt <- Event{
						Type: Name,
						ID:   auth.NewUUID(),
						Path: filePath,
					}
					initialStat = stat
				default:
					// wait before checking again
					time.Sleep(WAIT)
				}
			}
		}
	}
	// start watcher
	go func() {
		log.Printf("[INFO] monitoring %s...", baseName)
		watcher()
	}()
	return evt
}

// watch for changes in a directory
func watchDir(dirPath string, stop chan bool) chan Event {
	// get initial slice of file and subdirectories
	initialItems, err := os.ReadDir(dirPath)
	if err != nil {
		log.Printf("[ERROR] failed to ready directory contents: %v", err)
		return nil
	}

	// add initial items to context
	dirCtx := NewDirCtx()
	// NOTE: need to make sure these are added to the service
	// if they're not already present! not watchDir's responsibility, though.
	dirCtx.AddItems(initialItems, dirPath)

	// event channel used by the event handler goroutine
	evt := make(chan Event)

	// watcher function
	var watcher = func() {
		for {
			select {
			case <-stop:
				log.Printf("[INFO] stopping monitor for %s...", filepath.Base(dirPath))
				return
			default:
				currItems, err := os.ReadDir(dirPath)
				if err != nil {
					log.Printf("[ERROR] failed to read directory: %v", err)
					return
				}
				switch {
				// directory was deleted
				case err == os.ErrExist:
					evt <- Event{
						Type: Delete,
						Path: dirPath,
					}
					close(evt)
					return
				// item(s) were deleted
				case len(currItems) < len(initialItems):
					diffs := dirCtx.AddItems(currItems, dirPath) // get list of deleted items
					evt <- Event{
						Type:  Delete,
						Path:  dirPath,
						Items: diffs,
					}
					initialItems = currItems
				// item(s) were added
				case len(currItems) > len(initialItems):
					diffs := dirCtx.AddItems(currItems, dirPath) // get list of removed items
					evt <- Event{
						Type:  Add,
						Path:  dirPath,
						Items: diffs,
					}
					initialItems = currItems
				}
				// TODO: other directory changes?
			}
		}
	}

	// start watcher
	go func() {
		log.Printf("[INFO] monitoring %s...", filepath.Base(dirPath))
		watcher()
	}()

	return evt
}

// add all files and directories under the given path
// (assumed to be a root directory) to the monitoring instance
func watchAll(path string, m *Monitor) error {
	log.Printf("[INFO] adding watchers for all files and directories under %s ...", path)
	err := filepath.Walk(path, func(itemPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if err := m.WatchItem(itemPath); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk directory: %v", err)
	}
	log.Printf("[INFO] monitor is running. watching %d items", len(m.Events))
	return nil
}
