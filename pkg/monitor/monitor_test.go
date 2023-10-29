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
func testListener(t *testing.T, path string, stopMonitor chan bool, stopListener chan bool) {
	go func() {
		log.Print("listening for events...")
		fileChan := watchFile(path, stopMonitor)
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

// starts a new testListner for a given file.
// returns a monitor shutdown channel and a listener shut down channel
func NewTestListener(t *testing.T, path string) (chan bool, chan bool) {
	stopMonitor := make(chan bool)
	stopListener := make(chan bool)
	testListener(t, path, stopMonitor, stopListener)
	return stopMonitor, stopListener
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
	MutateFile(t, file)

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

func TestMonitorOneFileWithMultipleChanges(t *testing.T) {
	fn := filepath.Join(GetTestingDir(), "tmp.txt")

	file, err := MakeTmpTxtFile(fn, RandInt(1000))
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// listen for events from file monitor
	shutDown := make(chan bool)
	stopListener := make(chan bool)
	testListener(t, file.Path, shutDown, stopListener)

	// wait for listener to start
	time.Sleep(time.Second)

	// alter the file a bunch of times to generate detections
	log.Print("altering test file...")

	for k := 0; k < 10; k++ {
		MutateFile(t, file)
		time.Sleep(time.Millisecond * 500)
	}

	// wait for the listener goroutine to receive the event
	time.Sleep(time.Second)

	// shutdown monitoring thread
	log.Print("shutting down monitoring and listening threads...")
	shutDown <- true
	stopListener <- true

	// clean up
	if err := Clean(t, GetTestingDir()); err != nil {
		log.Fatal(err)
	}
}

func TestMonitorOneFileWithDifferentEvents(t *testing.T) {
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
	MutateFile(t, file)

	time.Sleep(time.Second)

	log.Print("deleting file...")
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

type OffSwitches struct {
	StopMonitor  chan bool
	StopListener chan bool
}

func TestMonitorWatchAll(t *testing.T) {
	tmp := MakeTmpDirs(t)

	// new monitor
	monitor := NewMonitor(tmp.Path)
	if err := monitor.Start(monitor.Path); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// get the files to monitor
	files, err := tmp.GetFiles()
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// create tmp listeners for each file monitor
	offSwitches := make([]*OffSwitches, 0, len(files))
	for i := 0; i < len(files); i++ {
		stopMonitor, stopListener := NewTestListener(t, files[i].Path)
		off := &OffSwitches{
			StopMonitor:  stopMonitor,
			StopListener: stopListener,
		}
		offSwitches = append(offSwitches, off)
	}

	// alter a bunch of the files at random
	for _, file := range files {
		choice := RandInt(2)
		switch choice {
		case 1:
			log.Print("altering file...")
			MutateFile(t, file)
		case 2:
			continue
		}
	}

	// stop all the listeners and monitors
	for _, off := range offSwitches {
		off.StopListener <- true
		off.StopMonitor <- true
	}

	if err := Clean(t, GetTestingDir()); err != nil {
		log.Fatal(err)
	}
}
