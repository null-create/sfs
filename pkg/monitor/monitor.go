package monitor

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/logger"
)

/*
this is the file for the background event-listener daemon.

this will listen for events like a file being saved within the client's drive, which will then
automatically start a new sync index operation. whether the user wants to automatically sync or not
should be a setting, but the daemon will automatically make a new sync index with each file or directory
modification.

should also have a mechanism to interrupt a sync operation if a new event occurs.
*/

// arbitrary* wait times between checks (*after some hand tuning)
const (
	WAIT        = time.Millisecond * 500 // wait duration after checks with no changes
	WAIT_LONGER = time.Second            // wait duration after checks with changes
)

type Watcher func(string, chan bool) chan Event

type Monitor struct {
	mu sync.Mutex // guards

	// logger for monitor
	log *logger.Logger

	// map of channels to active watchers.
	// key is the absolute path, value is the channel to the watchFile()
	// or watchDir() goroutine associated with that file or directory
	//
	// key = item path, val is Event channel
	Events map[string]chan Event

	// active watchers
	// key is the items absolute path, value is the watcher function instance.
	Watchers map[string]Watcher

	// map of channels to active watchers that will shut down the watcher
	// goroutine when set to true.
	//
	// key = item path, val is chan bool
	OffSwitches map[string]chan bool
}

func NewMonitor(drvRoot string) *Monitor {
	return &Monitor{
		log:         logger.NewLogger("Monitor", "None"),
		Events:      make(map[string]chan Event),
		Watchers:    make(map[string]Watcher),
		OffSwitches: make(map[string]chan bool),
	}
}

// see if an event channel exists for a given filepath.
func (m *Monitor) IsMonitored(path string) bool {
	if _, exists := m.Watchers[path]; exists {
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
		m.log.Error(fmt.Sprintf("failed to get stats for path %s: %v", path, err))
		return false
	}
	return true
}

// is this item a directory?
func (m *Monitor) IsDir(path string) (bool, error) {
	if stat, err := os.Stat(path); err == nil {
		return stat.IsDir(), nil
	} else {
		return false, fmt.Errorf("failed to get stats for %v: %v", filepath.Base(path), err)
	}
}

// add a new watcher function instance to the monitor.
func (m *Monitor) AddWatcher(path string, watcher Watcher) {
	if !m.IsMonitored(path) {
		m.Watchers[path] = watcher
	}
}

// start a watcher goroutine for a given path.
func (m *Monitor) StartWatcher(path string, stop chan bool) {
	if m.IsMonitored(path) {
		evts := m.Watchers[path](path, stop)
		m.Events[path] = evts
	}
}

// add a file to the events map and create a new monitoring
// goroutine. will need a corresponding events handler on the client end.
// will be a no-op if the given path is already being monitored, or if the
// supplied path points to a directory.
func (m *Monitor) Watch(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.Exists(path) {
		return fmt.Errorf("'%s' does not exist", filepath.Base(path))
	}
	if !m.IsMonitored(path) {
		isdir, err := m.IsDir(path)
		if err != nil {
			return err
		}
		// NOTE: monitoring directories is too expensive for the time being.
		// os.ReadDir() took a lot of CPU, especially when
		// called in a frequent operation loop. for now we're sticking
		// with monitoring files exclusively (or at least until we can find a
		// more efficient way to monitor directories.)
		if !isdir {
			stop := make(chan bool)
			m.OffSwitches[path] = stop
			m.AddWatcher(path, watch)
			m.StartWatcher(path, stop)
			m.log.Log(logger.INFO, fmt.Sprintf("monitoring %s...", filepath.Base(path)))
		}
	}
	return nil
}

// get an event listener channel for a given file.
// returns nil no listener channel is found.
func (m *Monitor) GetEventChan(path string) chan Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	if evtChan, exists := m.Events[path]; exists {
		return evtChan
	}
	m.log.Error(fmt.Sprintf("event channel for '%s' not found", filepath.Base(path)))
	return nil
}

// get an off switch for a given monitoring goroutine.
// off switches, when set to true, will shut down the monitoring process.
// returns nil if no off switch is available.
func (m *Monitor) GetOffSwitch(path string) chan bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if offSwitch, exists := m.OffSwitches[path]; exists {
		return offSwitch
	}
	m.log.Error(
		fmt.Sprintf("off switch not found for '%s' monitoring goroutine",
			filepath.Base(path),
		),
	)
	return nil
}

// close a watcher function and event channel for a given item.
// will be a no-op if the file is not registered.
func (m *Monitor) StopWatching(path string) {
	if m.IsMonitored(path) {
		m.Watchers[path] = nil
		delete(m.OffSwitches, path)
		delete(m.Events, path)
		delete(m.Watchers, path)
		m.log.Log(logger.INFO, fmt.Sprintf("'%s' is no longer being monitored", filepath.Base(path)))
	}
}

// shutdown all active monitoring threads
func (m *Monitor) ShutDown() {
	if len(m.Watchers) == 0 {
		return
	}
	m.log.Info(
		fmt.Sprintf("shutting down %d active monitoring threads...", len(m.Watchers)),
	)
	// the "graceful" way. blocks and is really slow.
	// for path := range m.OffSwitches {
	// 	m.OffSwitches[path] <- true
	// }
	// the "just erase it" way
	for key := range m.Watchers {
		m.Watchers[key] = nil
	}
}

// creates a new monitor goroutine for a given file or directory.
// returns a channel that sends events to the listener for handling
func watch(filePath string, stop chan bool) chan Event {
	log := logger.NewLogger("FILE_WATCHER", auth.NewUUID())

	// base file name for easier output reading
	baseName := filepath.Base(filePath)

	// event channel to pass file events to the event handler
	evtChan := make(chan Event)

	// get initial info for the file
	initialStat, err := os.Stat(filePath)
	if err != nil {
		log.Error(
			fmt.Sprintf("failed to get initial info for %s: %v - unable to monitor",
				filepath.Base(filePath), err,
			),
		)
		return nil
	}

	// dedicated watcher function
	var watcher = func() {
		for {
			select {
			case <-stop:
				log.Log(logger.INFO, fmt.Sprintf("shutting down monitoring for '%s'...", baseName))
				close(evtChan)
				return
			default:
				// TODO:
				// - examine performance around this os.Stat() call.
				//   it will be called A LOT of if we're watching hundreds or
				//   thousands of files across as many goroutines.
				//
				// - examine how it is implemented and see if we can hack a smaller DIY version, if necessary.
				//   also examine how fs-notify is implemented to get a better understanding of
				//   reading files and file systems for changes.
				//
				// - anything that needs to be changed should match how we're using the os.Stat(),
				//   and should only be concerned with adding an additional "backend," rather than
				//   changing anything on either side of this os.Stat() call.
				stat, err := os.Stat(filePath)
				if err != nil && err != os.ErrNotExist {
					log.Log(logger.INFO, fmt.Sprintf("%v - stopping monitoring for '%s'...", err, baseName))
					evtChan <- Event{
						IType: "File",
						Etype: Error,
						ID:    auth.NewUUID(),
						Path:  filePath,
					}
					close(evtChan)
					return
				}
				switch {
				// file deletion
				case err == os.ErrNotExist:
					log.Log(logger.INFO, fmt.Sprintf("'%s' was deleted. stopping monitoring.", baseName))
					evtChan <- Event{
						IType: "File",
						Etype: Delete,
						ID:    auth.NewUUID(),
						Path:  filePath,
					}
					close(evtChan)
					return
				// file size change
				case stat.Size() != initialStat.Size():
					log.Log(logger.INFO, fmt.Sprintf(
						"size change detected: %f kb -> %f kb | path: %s",
						float32(initialStat.Size()/1000), float32(stat.Size()/1000), filePath),
					)
					evtChan <- Event{
						IType: "File",
						Etype: Size,
						ID:    auth.NewUUID(),
						Path:  filePath,
					}
					initialStat = stat
				// file modification time change
				case stat.ModTime() != initialStat.ModTime():
					log.Log(logger.INFO, fmt.Sprintf("mod time change detected: %v -> %v", initialStat.ModTime(), stat.ModTime()))
					evtChan <- Event{
						IType: "File",
						Etype: ModTime,
						ID:    auth.NewUUID(),
						Path:  filePath,
					}
					initialStat = stat
				// file mode change
				case stat.Mode() != initialStat.Mode():
					log.Log(logger.INFO, (fmt.Sprintf("mode change detected: %v -> %v", initialStat.Mode(), stat.Mode())))
					evtChan <- Event{
						IType: "File",
						Etype: Mode,
						ID:    auth.NewUUID(),
						Path:  filePath,
					}
					initialStat = stat
				// file name change
				case stat.Name() != initialStat.Name():
					log.Log(logger.INFO, fmt.Sprintf("file name change detected: %v -> %v", initialStat.Name(), stat.Name()))
					evtChan <- Event{
						IType: "File",
						Etype: Name,
						ID:    auth.NewUUID(),
						Path:  filePath,
					}
					initialStat = stat
				default:
					// wait before checking again
					// TODO: experiment with longer wait times with non-buffered events.
					// want to try strike a balance between frequency of
					// checking vs faster checks with event bufferring.
					time.Sleep(WAIT)
				}
			}
		}
	}

	// start watcher
	go watcher()

	return evtChan
}

// alternative implementation using fsnotify. still experimental.
func watchfsn(filePath string, stop chan bool) chan Event {
	log := logger.NewLogger("FILE_WATCHER", auth.NewUUID())

	// base file name for easier output reading
	baseName := filepath.Base(filePath)

	// event channel to pass file events to the event handler
	evtChan := make(chan Event)

	// setup watcher
	w, err := fsnotify.NewBufferedWatcher(0)
	if err != nil {
		log.Error(err.Error())
	}

	// event loop
	go func() {
		for {
			select {
			case <-stop:
				log.Info("stopping watcher for " + baseName)
				w.Close()
				close(evtChan)
				return
			case event, ok := <-w.Events:
				if !ok {
					log.Error("failed to receive events for " + baseName)
					w.Close()
					close(evtChan)
					return
				}
				switch {
				case event.Has(fsnotify.Write):
					log.Log(logger.INFO, "writing change detected for "+baseName)
					evtChan <- Event{
						IType: "File",
						Etype: Size,
						ID:    auth.NewUUID(),
						Path:  filePath,
					}
				case event.Has(fsnotify.Chmod):
					log.Log(logger.INFO, "mode change detected for "+baseName)
					evtChan <- Event{
						IType: "File",
						Etype: Mode,
						ID:    auth.NewUUID(),
						Path:  filePath,
					}
				case event.Has(fsnotify.Rename):
					log.Log(logger.INFO, "file name change detected for "+baseName)
					evtChan <- Event{
						IType: "File",
						Etype: Name,
						ID:    auth.NewUUID(),
						Path:  filePath,
					}
				case event.Has(fsnotify.Remove):
					log.Log(logger.INFO, fmt.Sprintf("file '%s' removed", baseName))
					evtChan <- Event{
						IType: "File",
						Etype: Delete,
						ID:    auth.NewUUID(),
						Path:  filePath,
					}
					w.Close()
					close(evtChan)
					return
				default:
					time.Sleep(WAIT)
				}
			case err, ok := <-w.Errors:
				if !ok {
					w.Close()
					return
				}
				log.Error("error: " + err.Error())
			}
		}
	}()

	// Add the path to the item we want to monitor
	if err := w.Add(filePath); err != nil {
		log.Error("failed to add file to watcher: " + err.Error())
		w.Close()
		return nil
	}

	return evtChan
}

// add all files and directories under the given path
// (assumed to be a root directory) to the monitoring instance
func watchAll(path string, m *Monitor) error {
	m.log.Info(fmt.Sprintf("adding watchers for all files under %s ...", path))
	err := filepath.Walk(path, func(itemPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if err := m.Watch(itemPath); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk directory: %v", err)
	}
	return nil
}
