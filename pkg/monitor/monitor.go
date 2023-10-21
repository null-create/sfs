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
*/

const WAIT = time.Second * 60 // wait a  minute before checking file stat again after sending an event
const SHORT_WAIT = time.Second * 3

type Monitor struct {
	// path to the users drive root to monitor
	Path string

	// map of channels to active watchers.
	// key is the absolute file path, value is the channel to the watchFile() thread
	// associated with that file
	Events map[string]chan EventType

	// map of channels to active watchers that will shut down the watcher goroutine
	// when set to true.
	// key == file name, val is bool chan
	OffSwitches map[string]chan bool
}

func NewMonitor(drvRoot string) *Monitor {
	return &Monitor{
		Path: drvRoot,
	}
}

// creates a new monitor goroutine for a given file.
// returns a channel that sends events to the listener for handling
//
// TODO: add external shutdown capability (i.e. break for loop)
func watchFile(path string, stop chan bool) chan EventType {
	initialStat, err := os.Stat(path)
	if err != nil {
		log.Printf("[ERROR] failed to get file info for %s :%v\nunable to monitor", path, err)
		return nil
	}

	// event channel
	evt := make(chan EventType)

	go func() {
		for {
			stat, err := os.Stat(path)
			if err != nil {
				log.Printf("[WARN] failed to get file info for %s: %v\n stopping monitoring...", path, err)
				return
			}
			// TODO: maybe capture file state and info to match with from the user's files db.
			if stat.Size() != initialStat.Size() || stat.ModTime() != initialStat.ModTime() {
				log.Printf(
					fmt.Sprintf("[INFO] file change event -> [cur size: %d, prev size: %d] - [cur mod time: %s - prev mod time: %s]",
						stat.Size(), initialStat.Size(), stat.ModTime(), initialStat.ModTime()),
				)
				evt <- FileChange
				initialStat = stat

				// wait before checking again
				time.Sleep(WAIT)
			} else {
				// wait and try again.
				// TODO: customize wait times based on which of
				// the above conditions was true (i.e. don't read the file
				// too often if there's a lot of current activity with it)
				time.Sleep(SHORT_WAIT)
			}
			// shutdown signal received
			if <-stop {
				log.Printf("[INFO] stopping monitoring...")
				return
			}
		}
	}()

	return evt
}

// generate a series of channels to watchFile() goroutines and populate
// the m.Events map
func (m *Monitor) WatchFiles(dirpath string) error {
	entries, err := os.ReadDir(dirpath)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return fmt.Errorf("no files in %s", dirpath)
	}
	for _, entry := range entries {
		fp := filepath.Join(dirpath, entry.Name())
		if _, exists := m.Events[fp]; !exists {
			shutDown := make(chan bool)
			m.Events[fp] = watchFile(fp, shutDown)
		}
	}
	return nil
}

func (m *Monitor) GetFilePaths() []string {
	if len(m.Events) == 0 {
		log.Printf("[INFO] no files being monitored")
		return nil
	}
	files := make([]string, 0, len(m.Events))
	for fp, _ := range m.Events {
		files = append(files, fp)
	}
	return files
}

func (m *Monitor) exists(path string) bool {
	if _, exists := m.Events[path]; exists {
		return true
	}
	return false
}

func (m *Monitor) GetEventChan(path string) chan EventType {
	if evtChan, exists := m.Events[path]; exists {
		return evtChan
	}
	log.Printf("[ERROR] file (%s) event channel not found", filepath.Base(path))
	return nil
}

func (m *Monitor) NewChan(path string) {
	if !m.exists(path) {
		shutDown := make(chan bool)
		m.Events[path] = watchFile(path, shutDown)
	}
}

func (m *Monitor) CloseChan(path string) error {
	if evtChan, exists := m.Events[path]; exists {
		close(evtChan)
		delete(m.Events, path)
		return nil
	}
	return fmt.Errorf("file (%s) event channel not found", filepath.Base(path))
}
