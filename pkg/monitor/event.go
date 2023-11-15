package monitor

import (
	"fmt"
	"log"
	"time"
)

type EventType string

const (
	FileCreate EventType = "create"
	FileDelete EventType = "delete"
	FileChange EventType = "change"

	DirCreate EventType = "create"
	DirDelete EventType = "delete"
	DirChange EventType = "change"
)

type Event struct {
	// UUID of the event
	ID string
	// time of the event
	Time time.Time
	// type of file event, i.e. create, edit, or delete
	Type EventType
	// location of the file event (path to the file itself)
	Path string
}

func (e *Event) String() string {
	return fmt.Sprintf(
		"Event\n	id: %v\n	time: %v\n	type: %v\n	path: %v",
		e.ID, e.Time, e.Type, e.Path,
	)
}

// Elist is a buffer for file events in order to maximize
// synchronization operations between client and server
type EList []Event

// arbitrary threshold limit for Elists
const THRESHOLD = 10

type Events struct {
	threshold int   // buffer limit
	Total     int   // current total events
	StartSync bool  // flag to indicate whether a sync operation should start
	Events    EList // event object list
}

// new Events tracker. if buffered sync
// events will be delayed after THRESHOLD amount of events
// have been added to the EList buffer
func NewEvents(buffered bool) *Events {
	var threshold int
	if buffered {
		threshold = THRESHOLD
	} else {
		threshold = 1
	}
	return &Events{
		threshold: threshold,
		Events:    make(EList, 0),
	}
}

func (e *Events) Reset() {
	e.Events = make(EList, 0)
	e.StartSync = false
	e.Total = 0
}

func (e *Events) HasEvent(evt Event) bool {
	for _, evnt := range e.Events {
		if evt.ID == evnt.ID {
			return true
		}
	}
	return false
}

// add events until threshold is met.
// sets StartSync to true when threshold is met.
//
// additional events will be ignored until e.Reset() is called.
//
// if Events is buffered, then sync operations will be delayed
// until threshold is met, otherwise threshold is set to 1 by default
func (e *Events) AddEvent(evt Event) {
	if !e.HasEvent(evt) && e.Total+1 <= e.threshold {
		e.Events = append(e.Events, evt)
		e.Total += 1
		if e.Total == e.threshold {
			e.StartSync = true
		}
	} else {
		log.Printf("[WARNING] event list threshold met. event %s not added!", evt.ID)
	}
}
