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
	fileChan := watchFile(file.Path, shutDown)
	go func() {
		log.Print("[INFO] listening for events...")
		for {
			select {
			case <-fileChan:
				log.Printf("file event received: %v", file)
			}
		}
	}()
	time.Sleep(3 * time.Second)

	// alter the file to generate a detection
	log.Print("altering test file...")
	if err := file.Save([]byte(txtData)); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// verify detection occurred
	evt := <-fileChan
	if evt == "" {
		Fail(t, GetTestingDir(), fmt.Errorf("empty value received from listener"))
	}
	if evt != FileChange {
		Fail(t, GetTestingDir(), fmt.Errorf("wrong detection value received from listener: %v, wanted: %v", evt, FileChange))
	}
	log.Printf("file event received: %v", evt)

	// shut down the listener thread
	shutDown <- true
	close(shutDown)
	close(fileChan)

	// clean up
	if err := Clean(t, GetTestingDir()); err != nil {
		log.Fatal(err)
	}
}

func TestMonitorWithManyFiles(t *testing.T) {}
