package monitor

import (
	"fmt"
	"log"
)

type EventType string

// event enums
const (
	Add     EventType = "add"
	Create  EventType = "create"
	Delete  EventType = "delete"
	Remove  EventType = "remove"
	Change  EventType = "change"
	ModTime EventType = "modtime"
	Size    EventType = "size"
	Mode    EventType = "mode"
	Path    EventType = "path"
	Name    EventType = "name"
	Error   EventType = "error"
)

// an associated item (file or directory) for a given event
type EItem struct {
	name  string // item name
	path  string // item path
	itype string // item type (file or directory)
}

func (e *EItem) Name() string { return e.name }
func (e *EItem) Path() string { return e.path }
func (e *EItem) Kind() string { return e.itype }

type Event struct {
	ID    string    // UUID of the event
	Type  EventType // type of file event, i.e. create, edit, or delete
	Path  string    // location of the file event (path to the file itself)
	Items []EItem   // list of files or subdirectories in the directory that were added, created, or deleted
}

func (e *Event) ToString() string {
	return fmt.Sprintf(
		"item event \n(id=%s) -> type: %s | path: %s",
		e.ID, e.Type, e.Path,
	)
}

// Elist is a buffer for monitoring events.
type EList []Event

type Events struct {
	threshold int   // buffer limit
	Buffered  bool  // whether this event list is buffered
	Total     int   // current total events
	AtCap     bool  // flag to indicate whether we've reached the buffer limit
	Events    EList // event object list
}

// new Events tracker. if buffered sync
// events will be delayed after THRESHOLD amount of events
// have been added to the EList buffer
func NewEvents(buffered bool) *Events {
	var threshold int
	if buffered {
		threshold = MonCfgs.BufSize
	} else {
		threshold = 1
	}
	return &Events{
		threshold: threshold,
		Buffered:  buffered,
		Events:    make(EList, 0),
	}
}

func (e *Events) Reset() {
	e.Events = nil
	e.Events = make(EList, 0)
	e.AtCap = false
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
// sets e.AtCap to true when threshold is met.
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
			e.AtCap = true
		}
	} else {
		log.Printf("[WARNING] event list threshold met. event (id=%s) not added!", evt.ID)
	}
}

// returns a slice of file or directory paths to be used
// during sync operations
func (e *Events) GetPaths() ([]string, error) {
	if len(e.Events) == 0 {
		return nil, fmt.Errorf("event list is empty")
	}
	var paths []string
	for _, evt := range e.Events {
		paths = append(paths, evt.Path)
	}
	return paths, nil
}
