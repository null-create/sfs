package monitor

import "time"

type EventType string

const (
	FileChange EventType = "change"
	FileDelete EventType = "delete"
	FileCreate EventType = "create"
)

type Event struct {
	// type of file event, i.e. create, edit, or delete
	Type EventType `json:"type"`

	// location of the file event (path to the file itself)
	Path string `json:"path"`

	// time of the event
	Time time.Time `json:"time"`
}
