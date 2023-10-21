package monitor

import (
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

	// listen for events from file monitor
	shutDown := make(chan bool)
	fileChan := watchFile(file.Path, shutDown)
	go func() {
		log.Print("listening for events...")
		for {
			select {
			case <-fileChan:
				log.Print("file event received")
				time.Sleep(1 * time.Second)
			}
		}
	}()
	time.Sleep(2 * time.Second)

	// alter the file to generate a detection
	log.Print("altering test file...")
	if err := file.Save([]byte(txtData)); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// shutdown monitoring thread
	log.Print("shutting down monitoring thread...")
	shutDown <- true

	// clean up
	if err := Clean(t, GetTestingDir()); err != nil {
		log.Fatal(err)
	}
}

func TestMonitorWithManyFiles(t *testing.T) {}
