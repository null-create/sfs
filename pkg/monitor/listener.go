package monitor

import "log"

/*
This file works with the monitor to handle events and trigger
synchronization operations accordingly.
*/

type Listener struct {
	Path string // path to file thats being monitored
}

func NewListener(path string) *Listener {
	return &Listener{
		Path: path,
	}
}

// listen for file events given a newly created fileChan channel.
// ideally this function would be called soon after a new file monitor instance
// is created, and could be used as a dedicated listener.
//
// listen returns an event channel that passes the detected event to the caller
// for synchronization handling. Listen is basically a passive loop that filters
// events for the caller, rather than handling the events directly itself.
func (l *Listener) Listen(stop chan bool, fileChan chan EventType) (chan bool, chan EventType) {
	off := make(chan bool)      // off switch for this listener loop
	evt := make(chan EventType) // channel to pass Events out to sync services
	go func() {
		log.Print("[INFO] listening for file events...")
		for {
			select {
			case event := <-fileChan:
				switch event {
				case FileCreate:
					evt <- FileCreate
				case FileChange:
					evt <- FileChange
				case FileDelete:
					evt <- FileDelete
				}
			case <-stop:
				log.Print("[INFO] stopping listener...")
				return
			}
		}
	}()
	// channel to stop monitoring
	// (the stop parameter is this channel)
	return off, evt
}
