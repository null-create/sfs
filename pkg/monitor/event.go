package monitor

type EventType string

// event enums
const (
	Add     EventType = "add"
	Create  EventType = "create"
	Delete  EventType = "delete"
	Change  EventType = "change"
	ModTime EventType = "modtime"
	Size    EventType = "size"
	Mode    EventType = "mode"
	Path    EventType = "path"
	Name    EventType = "name"
	Error   EventType = "error"
)

// an event item (file or directory)
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
	IType string    // Item type (file or directory)
	Path  string    // location of the file event (path to the file itself)
	Etype EventType // Event type, i.e. create, edit, or delete
	Items []EItem   // list of event items (files or directories)
}

// Elist is a buffer for monitoring events.
type EList []Event

type Events struct {
	threshold int   // buffer limit
	buffered  bool  // whether this event list is buffered
	Total     int   // current total events
	atcap     bool  // flag to indicate whether we've reached the buffer limit
	Events    EList // event buffer
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
		buffered:  buffered,
		Events:    make(EList, 0),
	}
}

// Whether we've reached the buffer limit
func (e *Events) AtCap() bool { return e.atcap }

// reset events buffer and internal flags
func (e *Events) Reset() {
	e.Events = nil
	e.Events = make(EList, 0)
	e.atcap = false
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
			e.atcap = true
		}
	}
}
