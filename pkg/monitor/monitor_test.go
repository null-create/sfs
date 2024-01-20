package monitor

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/sfs/pkg/env"
)

// creates a new listener goroutine and checks received events
func testListener(t *testing.T, path string, stopMonitor chan bool, stopListener chan bool) {
	go func() {
		log.Print("listening for events...")
		fileChan := watch(path, stopMonitor)
		for {
			select {
			case evt := <-fileChan:
				switch evt.Type {
				case Change:
					log.Print("file change event received")
					assert.Equal(t, Change, evt.Type)
					assert.Equal(t, path, evt.Path)
				case Delete:
					log.Print("file delete event received")
					assert.Equal(t, Delete, evt.Type)
					assert.Equal(t, path, evt.Path)
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
	env.SetEnv(false)

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
	env.SetEnv(false)

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
	env.SetEnv(false)

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

// type OffSwitches struct {
// 	StopMonitor  chan bool
// 	StopListener chan bool
// }

func TestMonitorWatchAll(t *testing.T) {
	env.SetEnv(false)

	tmp := MakeTmpDirs(t)

	// initialize new monitor with watching goroutines
	// for all files under tmp. none of the watchers will have event
	// listeners, we just want to see if they all independently
	// detect file changes.
	monitor := NewMonitor(tmp.Path)
	if err := monitor.Start(monitor.Path); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// get the files to monitor
	files := tmp.GetFiles()
	if files == nil {
		Fail(t, GetTestingDir(), fmt.Errorf("no files found"))
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

	// stop all  monitors
	//
	// NOTE: off switches don't seem to be working?
	// test times out currently
	for _, off := range monitor.OffSwitches {
		off <- true
	}

	if err := Clean(t, GetTestingDir()); err != nil {
		log.Fatal(err)
	}
}
