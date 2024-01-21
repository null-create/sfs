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
func (m *Monitor) Exists(filePath string) bool {
	if _, exists := m.Events[filePath]; exists {
		return true
	}
	return false
}

// recursively builds watchers for all files in the directory
// and subdirectories
func (m *Monitor) Start(dirpath string) error {
	return watchAll(dirpath, m)
}

func (m *Monitor) IsDir(path string) (bool, error) {
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		return true, nil
	} else if err != nil {
		return false, err
	}
	return false, nil
}

// add a file or directory to the events map and create a new monitoring
// goroutine. will need a corresponding events handler. will be a no-op if the
// given path is already being monitored.
func (m *Monitor) WatchItem(path string) error {
	// make sure this item actually exists
	if _, err := os.Stat(path); err != nil && errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%s does not exist", filepath.Base(path))
	} else if err != nil {
		return err
	}
	// if not already being monitored
	if !m.Exists(path) {
		stop := make(chan bool)
		m.OffSwitches[path] = stop
		m.Events[path] = watch(path, stop)
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
		log.Print("[WARNING] no event channels available")
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
	if m.Exists(filePath) {
		m.OffSwitches[filePath] <- true // shut down monitoring thread before closing
		delete(m.OffSwitches, filePath)
		delete(m.Events, filePath)
		log.Printf("[INFO] file monitoring channel (%s) closed", filepath.Base(filePath))
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
func watch(path string, stop chan bool) chan Event {
	initialStat, err := os.Stat(path)
	if err != nil {
		log.Printf("[ERROR] failed to get initial info for %s: %v\nunable to monitor", filepath.Base(path), err)
		return nil
	}

	// event channel used by the event handler goroutine
	evt := make(chan Event)

	go func() {
		log.Printf("[INFO] monitoring %s...", filepath.Base(path))
		for {
			select {
			case <-stop:
				log.Printf("[INFO] shutting down monitoring for %s...", filepath.Base(path))
				close(evt)
				return
			default:
				stat, err := os.Stat(path)
				if err != nil && err != os.ErrNotExist {
					log.Printf("[ERROR] %v\nstopping monitoring for %s...", err, filepath.Base(path))
					close(evt)
					return
				}
				switch {
				// file deletion
				case err == os.ErrNotExist:
					log.Printf("[INFO] %s deleted. stopping monitoring...", filepath.Base(path))
					evt <- Event{
						Type: Change,
						ID:   auth.NewUUID(),
						Time: time.Now().UTC(),
						Path: path,
					}
					close(evt)
					return
				// file size change
				case stat.Size() != initialStat.Size():
					log.Printf("[INFO] size change detected: %f kb -> %f kb", float64(initialStat.Size()/1000), float64(stat.Size()/1000))
					evt <- Event{
						Type: Change,
						ID:   auth.NewUUID(),
						Time: time.Now().UTC(),
						Path: path,
					}
					initialStat = stat
				// file modification time change
				case stat.ModTime() != initialStat.ModTime():
					log.Printf("[INFO] mod time change detected: %v -> %v", initialStat.ModTime(), stat.ModTime())
					evt <- Event{
						Type: Change,
						ID:   auth.NewUUID(),
						Time: time.Now().UTC(),
						Path: path,
					}
					initialStat = stat
				default:
					// wait before checking again
					time.Sleep(WAIT)
				}
			}
		}
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
	log.Printf("[INFO] monitor is running. watching %d files", len(m.Events))
	return nil
}
