package monitor

import (
	"fmt"
	"log"
	"os"
)

/*
File for managing the sync doc file shared between event handlers
and client monitor instances
*/

func NewSD(path string) {
	file, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	_, err = file.Write([]byte("0"))
	if err != nil {
		log.Fatal(err)
	}
}

func (m *Monitor) trunc() {
	err := os.Truncate(m.SyncDoc, 0)
	if err != nil {
		log.Fatalf("[WARNING] failed to truncate sync doc: %v", err)
	}
}

// make the flag doc for event handlers to mark when
// a sync operation is supposed to happen.
func (m *Monitor) MakeSyncDoc() error {
	if _, err := os.Stat(m.SyncDoc); err != nil && os.IsNotExist(err) {
		if _, err2 := os.Create(m.SyncDoc); err2 != nil {
			return fmt.Errorf("failed to create sync doc: %v", err2)
		}
	}
	return nil
}

// delete the sync doc
func (m *Monitor) DeleteDoc() error {
	return os.Remove(m.SyncDoc)
}

// checks for whether the sync flag is set.
// if so, then a sync operation should take place.
func (m *Monitor) StartSync() bool {
	file, err := os.Open(m.SyncDoc)
	if err != nil {
		log.Printf("[ERROR] failed to open sync doc: %v", err)
		return false
	}
	defer file.Close()

	contents := make([]byte, 0)
	_, err = file.Read(contents)
	if err != nil {
		log.Printf("[ERROR] failed to open sync doc: %v", err)
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
		log.Fatalf("failed to open sync doc: %v", err)
	}
	defer file.Close()

	m.trunc()

	_, err = file.Write([]byte("0"))
	if err != nil {
		log.Fatalf("failed to update sync doc: %v", err)
	}
}
