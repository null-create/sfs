package monitor

import (
	"fmt"
	"log"
	"path/filepath"
	"testing"
	"time"
)

func TestMonitorWithOneFile(t *testing.T) {
	fn := filepath.Join(GetTestingDir(), "tmp.txt")

	file, err := MakeTmpTxtFile(fn, RandInt(1000))
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// start monitoring thread & wait for a few seconds
	log.Print("starting monitoring thread...")

	// listen for events from file monitor
	shutDown := make(chan bool)
	fileChan := watchFile(file.Path)
	go func() {
		log.Print("[INFO] listening for events...")
		for {
			select {
			case <-fileChan:
				log.Printf("file event received: %v", file)
			case <-shutDown:
				log.Printf("shutting down listener thread...")
				return
			}
		}
	}()
	time.Sleep(2 * time.Second)

	// alter the file to generate a detection
	log.Print("altering test file...")
	if err := file.Save([]byte(txtData)); err != nil {
		shutDown <- true
		close(shutDown)
		Fail(t, GetTestingDir(), err)
	}

	// verify detection occurred
	evt := <-fileChan
	if evt == "" {
		shutDown <- true
		close(shutDown)
		Fail(t, GetTestingDir(), fmt.Errorf("empty value received from listener"))
	}
	if evt != FileChange {
		shutDown <- true
		close(shutDown)
		Fail(t, GetTestingDir(), fmt.Errorf("wrong detection value received from listener: %v, wanted: %v", evt, FileChange))
	}

	// shut down the listener thread
	shutDown <- true
	close(shutDown)

	// clean up
	if err := Clean(t, GetTestingDir()); err != nil {
		log.Fatal(err)
	}
}

func TestMonitorWithManyFiles(t *testing.T) {}
