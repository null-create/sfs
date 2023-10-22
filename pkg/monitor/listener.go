package monitor

import "log"

/*
This file works with the monitor to handle events and trigger
synchronization operations accordingly.
*/

type Listener struct{}

func NewListener() *Listener { return &Listener{} }

// listen for file events given a newly created fileChan channel.
// ideally this function would be called soon after a new file monitor instance
// is created, and could be used as a dedicated listener.
func (l *Listener) Listen(stop chan bool, fileChan chan EventType) chan bool {
	off := make(chan bool)
	go func() {
		log.Print("[INFO] listening for file events...")
		for {
			select {
			case event := <-fileChan:
				switch event {
				// TODO: handle file events lol
				case FileCreate:
				case FileChange:
				case FileDelete:
				}
			case <-stop:
				log.Print("[INFO] stopping listener...")
				return
			}
		}
	}()
	// channel to stop monitoring
	// (the stop parameter is this channel)
	return off
}
