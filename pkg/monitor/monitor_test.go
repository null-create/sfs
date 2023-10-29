package monitor

import (
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
)

// creates a new listener goroutine and checks received events
func testListener(t *testing.T, path string, shutDown chan bool, stopListener chan bool) {
	go func() {
		log.Print("listening for events...")
		fileChan := watchFile(path, shutDown)
		for {
			select {
			case evt := <-fileChan:
				switch evt.Type {
				case FileChange:
					log.Print("file change event received")
					assert.Equal(t, FileChange, evt.Type)
					assert.Equal(t, path, evt.Path)
				case FileDelete:
					log.Print("file delete event received")
					assert.Equal(t, FileDelete, evt.Type)
					assert.Equal(t, path, evt.Path)
				default:
					log.Printf("unknown event type: %v", evt.Type)
				}
			case <-stopListener:
				log.Print("shutting down listener...")
				return
			default:
				continue
			}
		}
	}()
}

func TestMonitorWithOneFile(t *testing.T) {
	fn := filepath.Join(GetTestingDir(), "tmp.txt")

	file, err := MakeTmpTxtFile(fn, RandInt(1000))
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// listen for events from file monitor
	shutDown := make(chan bool)
	stopListener := make(chan bool)
	testListener(t, file.Path, shutDown, stopListener)

	time.Sleep(2 * time.Second)

	// alter the file to generate a detection
	log.Print("altering test file...")

	// make a huge string so we can hopefully
	// detect the change
	var data string
	for i := 0; i < 10000; i++ {
		data += txtData
	}
	if err := file.Save([]byte(data)); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// wait for the listener goroutine to receive the event
	time.Sleep(2 * time.Second)

	// shutdown monitoring thread
	log.Print("shutting down monitoring and listening threads...")
	shutDown <- true
	stopListener <- true

	// clean up
	if err := Clean(t, GetTestingDir()); err != nil {
		log.Fatal(err)
	}
}

func TestMonitorWithMultipleChanges(t *testing.T) {
	fn := filepath.Join(GetTestingDir(), "tmp.txt")

	file, err := MakeTmpTxtFile(fn, RandInt(1000))
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// listen for events from file monitor
	shutDown := make(chan bool)
	stopListener := make(chan bool)
	testListener(t, file.Path, shutDown, stopListener)

	time.Sleep(2 * time.Second)

	// alter the file a bunch of times to generate detections
	log.Print("altering test file...")

	for k := 0; k < 10; k++ {
		// make a huge string so we can hopefully
		// detect the change
		var data string
		for i := 0; i < 1000; i++ {
			data += txtData
		}
		if err := file.Save([]byte(data)); err != nil {
			Fail(t, GetTestingDir(), err)
		}
	}

	// wait for the listener goroutine to receive the event
	time.Sleep(2 * time.Second)

	// shutdown monitoring thread
	log.Print("shutting down monitoring and listening threads...")
	shutDown <- true
	stopListener <- true

	// clean up
	if err := Clean(t, GetTestingDir()); err != nil {
		log.Fatal(err)
	}
}

func TestMonitorWithDifferentEvents(t *testing.T) {
	fn := filepath.Join(GetTestingDir(), "tmp.txt")

	file, err := MakeTmpTxtFile(fn, RandInt(1000))
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// listen for events from file monitor
	shutDown := make(chan bool)
	stopListener := make(chan bool)
	testListener(t, file.Path, shutDown, stopListener)

	time.Sleep(time.Second)

	log.Print("altering file...")
	// add a big string to detect the change
	var data string
	for i := 0; i < 10000; i++ {
		data += txtData
	}
	if err := file.Save([]byte(data)); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	time.Sleep(time.Second)

	log.Print("deleting file...")
	// delete file to generate a deletion event
	if err := os.Remove(file.Path); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// shutdown monitoring thread & clean up
	log.Print("shutting down monitoring and listening threads...")
	shutDown <- true
	stopListener <- true
	if err := Clean(t, GetTestingDir()); err != nil {
		log.Fatal(err)
	}
}
