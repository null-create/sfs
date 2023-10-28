package monitor

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

/*
this is the file for the background event-listener daemon.

this will listen for events like a file being saved within the client's drive, which will then
automatically start a new sync index operation. whether the user wants to automatically sync or not
should be a setting, but the daemon will automatically make a new sync index with each file or directory
modification.

should also have a mechanism to interrupt a sync operation if a new event occurs.

NOTE: a new watcher should be created whenever a new file is created on the server or client,
and removed when a file is deleted.

see: https://stackoverflow.com/questions/8270441/go-language-how-detect-file-changing
*/

const WAIT = time.Second * 10 // wait 10 seconds before checking file stat again after sending an event
const SHORT_WAIT = time.Second * 3

type Monitor struct {
	// path to the users drive root to monitor
	Path string

	// map of channels to active listeners.
	// key is the absolute file path, value is the channel to the watchFile() thread
	// associated with that file
	//
	// key = file path, val is EventType channel
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

// creates a new monitor goroutine for a given file.
// returns a channel that sends events to the listener for handling
func watchFile(path string, stop chan bool) chan Event {
	initialStat, err := os.Stat(path)
	if err != nil {
		log.Printf("[ERROR] failed to get file info for %s :%v\nunable to monitor", path, err)
		return nil
	}

	// event channel
	evt := make(chan Event)

	go func() {
		log.Print("[INFO] starting monitoring...")
		for {
			stat, err := os.Stat(path)
			log.Printf("stat: %s", path)
			if err != nil && err != os.ErrNotExist {
				log.Printf("[ERROR] failed to get file info: %v\nstopping monitoring...", err)
				close(evt)
				return
			}
			// events
			switch {
			// file deleted
			case err == os.ErrNotExist:
				log.Printf("[INFO] file %s deleted. shuttown down monitoring...", path)
				evt <- Event{
					Type: FileDelete,
					Time: time.Now().UTC(),
					Path: path,
				}
				close(evt)
				return
			// file size change
			case stat.Size() != initialStat.Size():
				log.Printf("[INFO] file size change detected: %d -> %d", stat.Size(), initialStat.Size())
				evt <- Event{
					Type: FileChange,
					Time: time.Now().UTC(),
					Path: path,
				}
				initialStat = stat
			// file mod time change
			case stat.ModTime() != initialStat.ModTime():
				log.Printf("[INFO] file mod time change detected: %v -> %v", stat.ModTime(), initialStat.ModTime())
				evt <- Event{
					Type: FileChange,
					Time: time.Now().UTC(),
					Path: path,
				}
				initialStat = stat
			// stop monitoring
			case <-stop:
				log.Printf("[INFO] shutting down monitoring...")
				close(evt)
				return
			default:
				continue
			}
		}
	}()

	return evt
}

func watchAll(path string, m *Monitor) error {
	log.Printf("[INFO] adding watchers for all files under %s ...", path)
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// make sure this a file
		if stat, err := os.Stat(filePath); !stat.IsDir() {
			m.WatchFile(filePath)
		} else if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk directory: %v", err)
	}
	return nil
}

// recursively builds watchers for all files in the directory
// and subdirectories
func (m *Monitor) WatchAll(dirpath string) error {
	entries, err := os.ReadDir(dirpath)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		log.Printf("no files or subdirectories in %s", dirpath)
		return nil
	}
	return watchAll(dirpath, m)
}

// add a file to the events map
func (m *Monitor) WatchFile(filePath string) {
	if _, exists := m.Events[filePath]; !exists {
		stop := make(chan bool)
		m.OffSwitches[filePath] = stop
		m.Events[filePath] = watchFile(filePath, stop)
	}
}

func (m *Monitor) exists(path string) bool {
	if _, exists := m.Events[path]; exists {
		return true
	}
	return false
}

// get an event listener channel for a given file
func (m *Monitor) GetEventChan(path string) chan Event {
	if evtChan, exists := m.Events[path]; exists {
		return evtChan
	}
	log.Printf("[ERROR] file (%s) event channel not found", filepath.Base(path))
	return nil
}

// get an event listener off switch for a given file
func (m *Monitor) GetOffSwitch(path string) chan bool {
	if offSwitch, exists := m.OffSwitches[path]; exists {
		return offSwitch
	}
	log.Printf("[ERROR] off switch not found for file (%s) monitoring goroutine", filepath.Base(path))
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

// create a watcher thread for a given file
func (m *Monitor) NewWatcher(path string) {
	if !m.exists(path) {
		shutDown := make(chan bool)
		m.OffSwitches[path] = shutDown
		m.Events[path] = watchFile(path, shutDown)
		log.Printf("[INFO] file (%s) watcher created", filepath.Base(path))
	}
}

// close a listener channel for a given file
func (m *Monitor) CloseChan(filePath string) error {
	if m.exists(filePath) {
		m.OffSwitches[filePath] <- true // shut down monitoring thread before closing
		delete(m.OffSwitches, filePath)
		delete(m.Events, filePath)
		log.Printf("[INFO] file channel (%s) closed", filePath)
		return nil
	}
	return fmt.Errorf("file (%s) event channel not found", filepath.Base(filePath))
}

// shutdown all active monitoring threads
func (m *Monitor) ShutDown() error {
	if len(m.OffSwitches) == 0 {
		return fmt.Errorf("no event off channels available")
	}
	paths := m.GetPaths()
	if paths == nil {
		return fmt.Errorf("no paths available")
	}
	for _, path := range paths {
		m.OffSwitches[path] <- true
	}
	return nil
}
