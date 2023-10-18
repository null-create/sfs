package monitor

import (
	"log"
	"os"
	"time"
)

/*
this is the file for the background event-listener daemon.

this will listen for events like a file being saved within the client's drive, which will then
automatically start a new sync index operation. whether the user wants to automatically sync or not
should be a setting, but the daemon will automatically make a new sync index with each file or directory
modification.

should also have a mechanism to interrupt a sync operation if a new event occurs.
*/

const MAXStatRetries = 10
const WAIT = time.Second * 60 // wait a  minute before checking file stat again after sending an event

type Monitor struct {
	// path to the users drive root to monitor
	Path string

	// map of channels to active watchers (channel type TBD).
	// key is the file, value is the channel
	Events map[string]<-chan Event
}

func NewMonitor(drvRoot string) *Monitor {
	return &Monitor{
		Path: drvRoot,
	}
}

// creates a new monitor goroutine for a given file.
// returns a channel that sends events to the listener for handling
func watchFile(path string) <-chan Event {
	initialStat, err := os.Stat(path)
	if err != nil {
		log.Printf("[ERROR] failed to get file info: %v", err)
		return nil
	}

	var retries int
	for {
		stat, err := os.Stat(path)
		if err != nil {
			log.Printf("[WARN] failed to get file info for %s: %v", path, err)
			retries += 1
			if retries == MAXStatRetries {
				log.Printf("[ERROR] unable to get file info for %s after %d retries. cancelling monitoring...", path, MAXStatRetries)
				return nil
			}
		}
		// TODO:
		// send notification via a channel using an "event" struct.
		// maybe capture file state and info to match with from the user's files db.
		// also shouldn't break loop -- that's just a placeholder.
		// also we need a way to reset initialStat to stat
		if stat.Size() != initialStat.Size() {
			// add event channel here

			// reset
			initialStat = stat
			time.Sleep(WAIT)
		} else if stat.ModTime() != initialStat.ModTime() {
			// add event channel here

			// reset
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
}

// TODO: function that builds watchFile goroutines for all files in a user's drive root
// directory. this will be the new constructor for Monitor.
