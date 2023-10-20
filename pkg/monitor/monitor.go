package monitor

import (
	"log"
	"os"
	"time"

	"github.com/pkg/errors"
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

type Monitor struct {
	// path to the users drive root to monitor
	Path string

	// map of channels to active watchers.
	// key is the file ID, value is the channel to the watchFile() thread
	// associated with that file
	Events map[string]chan EventType
}

func NewMonitor(drvRoot string) *Monitor {
	return &Monitor{
		Path: drvRoot,
	}
}

// creates a new monitor goroutine for a given file.
// returns a channel that sends events to the listener for handling
func watchFile(path string) chan EventType {
	initialStat, err := os.Stat(path)
	if err != nil {
		log.Printf("[ERROR] failed to get file info: %v", err)
		return nil
	} else if errors.Is(err, os.ErrNotExist) {
		log.Printf("[ERROR] file does not exist: %v", path)
		return nil
	}

	// event channel
	evt := make(chan EventType)

	go func() {
		for {
			stat, err := os.Stat(path)
			if err != nil {
				log.Printf("[WARN] failed to get file info for %s: %v", path, err)
				return
			}
			// TODO:
			// maybe capture file state and info to match with from the user's files db.
			if stat.Size() != initialStat.Size() {
				evt <- FileChange
				initialStat = stat

				time.Sleep(WAIT)
			} else if stat.ModTime() != initialStat.ModTime() {
				evt <- FileChange
				initialStat = stat

				time.Sleep(WAIT)
			} else {
				// wait and try again.
				// TODO: customize wait times based on which of
				// the above conditions was true (i.e. don't read the file
				// too often if there's a lot of current activity with it)
				time.Sleep(1 * time.Second)
			}
		}
	}()

	return evt
}

// TODO: function that builds watchFile goroutines for all files in a user's drive root
// directory. this will be the new constructor for Monitor.
