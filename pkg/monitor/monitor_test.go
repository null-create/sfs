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
		log.Printf("[TEST] listening for %s events...", filepath.Base(path))
		fileChan := watchFile(path, stopMonitor)
		for {
			select {
			case evt := <-fileChan:
				switch evt.Type {
				case Change:
					log.Print("[TEST] file change event received")
					assert.Equal(t, Change, evt.Type)
					assert.Equal(t, path, evt.Path)
				case Delete:
					log.Print("[TEST] file delete event received")
					assert.Equal(t, Delete, evt.Type)
					assert.Equal(t, path, evt.Path)
				}
			case <-stopListener:
				log.Print("[TEST] shutting down listener...")
				return
			default:
				continue
			}
		}
	}()
}

func testMonitorListener(t *testing.T, path string, stopMonitor chan bool, stopListener chan bool) {
	go func() {
		log.Print("[TEST] monitoring directory: " + filepath.Base(path))
		dirChan := watchDir(path, stopMonitor)
		for {
			select {
			case evt := <-dirChan:
				switch evt.Type {
				case Add:
					log.Print("[TEST] add event detected: " + evt.Path)
					var items string
					for _, evt := range evt.Items {
						items += evt.name + "\n"
					}
					log.Printf("[TEST] items: " + items)
				default:
					continue
				}
			case <-stopListener:
				log.Print("[TEST] stopping listener...")
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

func NewTestMonitorListener(t *testing.T, path string) (chan bool, chan bool) {
	stopMonitor := make(chan bool)
	stopListener := make(chan bool)
	testMonitorListener(t, path, stopMonitor, stopListener)
	return stopMonitor, stopListener
}

func TestMonitorWithOneFile(t *testing.T) {
	env.SetEnv(false)

	file, err := MakeTmpTxtFile(filepath.Join(GetTestingDir(), "tmp.txt"), RandInt(1000))
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

	file, err := MakeTmpTxtFile(filepath.Join(GetTestingDir(), "tmp.txt"), RandInt(1000))
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

	file, err := MakeTmpTxtFile(filepath.Join(GetTestingDir(), "tmp.txt"), RandInt(1000))
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

func TestMonitorDirectory(t *testing.T) {
	env.SetEnv(false)

	// make temp files to monitor
	tmp := MakeTmpDirs(t)

	// initialize new monitor with watching goroutines
	// for all files under tmp. none of the watchers will have event
	// listeners, we just want to see if they all independently
	// detect file changes.
	monitor := NewMonitor(tmp.Path)
	if err := monitor.Start(tmp.Path); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// add a test listener for the temp directory
	stopListener := make(chan bool)
	testMonitorListener(t, tmp.Path, make(chan bool), stopListener)

	// make a new file in the temp directory and add to the monitor
	file, err := MakeTmpTxtFile(filepath.Join(tmp.Path, "new-thing.txt"), RandInt(500))
	if err != nil {
		monitor.ShutDown()
		stopListener <- true
		Fatal(t, err)
	}
	if err := monitor.WatchItem(file.Path); err != nil {
		monitor.ShutDown()
		stopListener <- true
		Fail(t, GetTestingDir(), err)
	}

	// shut down and clean up
	monitor.ShutDown()
	stopListener <- true
	if err := Clean(t, GetTestingDir()); err != nil {
		log.Fatal(err)
	}
}
