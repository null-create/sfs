package monitor

import (
	"log"
	"path/filepath"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
)

func testListener(fileChan chan EventType, stopListener chan bool) {
	log.Print("listening for events...")
	for {
		select {
		case evt := <-fileChan:
			switch evt {
			case FileChange:
				log.Print("file event received")
			}
		case <-stopListener:
			log.Print("shutting down listener...")
			return
		default:
			continue
		}
	}
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
	go func() {
		log.Print("listening for events...")
		fileChan := watchFile(file.Path, shutDown)
		for {
			select {
			case evt := <-fileChan:
				log.Print("file event received")
				assert.Equal(t, FileChange, evt.Type)
				assert.Equal(t, file.Path, evt.Path)
			case <-stopListener:
				log.Print("shutting down listener...")
				return
			default:
				continue
			}
		}
	}()

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

func TestMonitorWithManyFiles(t *testing.T) {}
