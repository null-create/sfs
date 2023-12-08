package monitor

import (
	"log"
	"os"
)

/*
File for managing the sync doc file shared between event handlers
and client monitor instances
*/

func (m *Monitor) trunc() {
	err := os.Truncate(m.SyncDoc, 0)
	if err != nil {
		log.Fatalf("[WARNING] failed to truncate sync doc: %v", err)
	}
}

func (m *Monitor) IsReady() bool {
	file, err := os.Open(m.SyncDoc)
	if err != nil {
		log.Printf("[WARNING] failed to open sync doc: %v", err)
		return false
	}
	defer file.Close()

	contents := make([]byte, 0)
	_, err = file.Read(contents)
	if err != nil {
		log.Printf("[WARNING] failed to open sync doc: %v", err)
		return false
	}
	return string(contents) == "1"
}

// sets the doc val to 1 for an event handler to indicate that
// a sync event should start
func (m *Monitor) SetDoc() {
	file, err := os.Open(m.SyncDoc)
	if err != nil {
		return
	}
	defer file.Close()

	m.trunc()

	_, err = file.Write([]byte("1"))
	if err != nil {
		log.Print("[WARNING] failed to update sync doc: ", err)
		return
	}
}

func (m *Monitor) ResetDoc() {
	file, err := os.Open(m.SyncDoc)
	if err != nil {
		log.Print("[WARNING] failed to open sync doc: ", err)
		return
	}
	defer file.Close()

	m.trunc()

	_, err = file.Write([]byte("0"))
	if err != nil {
		log.Print("[WARNING] failed to update sync doc: ", err)
	}
}
