package monitor

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
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
	// mutex lock to protect the sync document
	mu sync.Mutex

	// path to the users drive root to monitor
	Path string

	// path to the .txt file containing the flag to
	// indicate whether the events handler has indicated that
	// a sync operation should be performed.
	SyncDoc string

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
	syncDoc := filepath.Join(drvRoot, ".sync.txt")
	return &Monitor{
		Path:        drvRoot,
		SyncDoc:     syncDoc,
		Events:      make(map[string]chan Event),
		OffSwitches: make(map[string]chan bool),
	}
}

func (m *Monitor) Exists(path string) bool {
	if _, exists := m.Events[path]; exists {
		return true
	}
	return false
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
			select {
			case <-stop:
				log.Printf("[INFO] shutting down monitoring...")
				close(evt)
				return
			default:
				stat, err := os.Stat(path)
				if err != nil && err != os.ErrNotExist {
					log.Printf("[ERROR] failed to get file info: %v\nstopping monitoring...", err)
					close(evt)
					return
				}
				switch {
				case err == os.ErrNotExist:
					log.Printf("[INFO] file %s deleted. shutting down monitoring...", path)
					evt <- Event{
						Type: FileDelete,
						ID:   auth.NewUUID(),
						Time: time.Now().UTC(),
						Path: path,
					}
					close(evt)
					return
				case stat.Size() != initialStat.Size() || stat.ModTime() != initialStat.ModTime():
					log.Printf("[INFO] file size change detected: %f kb -> %f kb", float64(initialStat.Size()/1000), float64(stat.Size()/1000))
					evt <- Event{
						Type: FileChange,
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

// add all files under the given path (assumed to be a root directory)
// to the monitoring instance
func watchAll(path string, m *Monitor) error {
	log.Printf("[INFO] adding watchers for all files under %s ...", path)
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// m.WatchFile handles whether this is a directory or a file.
		// we just don't want to miss anything.
		m.WatchFile(filePath)
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk directory: %v", err)
	}
	// make sure  we can shut everything down if need be
	if len(m.Events) != len(m.OffSwitches) {
		return fmt.Errorf(
			"failed to add off switches for all monitors. \nevts = %d, off-switches = %d",
			len(m.Events), len(m.OffSwitches),
		)
	}
	log.Printf("[INFO] monitor is running. watching (%d) files", len(m.Events))
	return nil
}

// recursively builds watchers for all files in the directory
// and subdirectories
func (m *Monitor) Start(dirpath string) error {
	log.Print("[INFO] starting monitor...")
	entries, err := os.ReadDir(dirpath)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		log.Printf("[WARNING] no files or subdirectories in %s", dirpath)
		return nil
	}
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

// make the tmp doc for event handlers to mark when
// a sync operation is supposed to happen.
func (m *Monitor) makeSyncDoc(fileID string, path string) error {
	fn := fmt.Sprintf(".sync-%s.txt", fileID)
	if _, err := os.Stat(filepath.Join(path, fn)); err != nil && os.IsNotExist(err) {
		if _, err2 := os.Create(fn); err2 != nil {
			return fmt.Errorf("failed to create sync doc: %v", err2)
		}
		return fmt.Errorf("failed to get file info: %v", err)
	}
	return nil
}

func (m *Monitor) remSyncDoc(path string) error {

	return nil
}

// add a file to the events map. will be a no-op if the
// given path is not a file path.
func (m *Monitor) WatchFile(filePath string) {
	if isDir, err := m.IsDir(filePath); isDir {
		log.Printf("[WARNING] path is a directory, not a file: %s", filePath)
		return
	} else if err != nil {
		log.Printf("[WARNING] failed to read file path: %s\nerr: %v", filePath, err)
		return
	}
	if !m.Exists(filePath) {
		stop := make(chan bool)
		m.OffSwitches[filePath] = stop
		m.Events[filePath] = watchFile(filePath, stop)
	}
}

// get an event listener channel for a given file
func (m *Monitor) GetEventChan(filePath string) chan Event {
	if evtChan, exists := m.Events[filePath]; exists {
		return evtChan
	}
	log.Printf("[ERROR] file (%s) event channel not found", filepath.Base(filePath))
	return nil
}

// get an event listener off switch for a given file
func (m *Monitor) GetOffSwitch(filePath string) chan bool {
	if offSwitch, exists := m.OffSwitches[filePath]; exists {
		return offSwitch
	}
	log.Printf("[ERROR] off switch not found for file (%s) monitoring goroutine", filepath.Base(filePath))
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

// close a listener channel for a given file
func (m *Monitor) CloseChan(filePath string) {
	if m.Exists(filePath) {
		m.OffSwitches[filePath] <- true // shut down monitoring thread before closing
		delete(m.OffSwitches, filePath)
		delete(m.Events, filePath)
		log.Printf("[INFO] file channel (%s) closed", filepath.Base(filePath))
	}
}

// shutdown all active monitoring threads
func (m *Monitor) ShutDown() error {
	if len(m.OffSwitches) == 0 {
		log.Printf("[INFO] no event off channels available. nothing to shutdown.")
		return nil
	}
	paths := m.GetPaths()
	if paths == nil {
		return fmt.Errorf("no paths available")
	}
	log.Print("[INFO] shutting down all active monitoring threads...")
	for _, path := range paths {
		m.OffSwitches[path] <- true
	}
	// delete sync doc

	return nil
}
