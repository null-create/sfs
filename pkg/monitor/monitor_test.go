package monitor

import (
	"fmt"
	"log"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestMonitorWithOneFile(t *testing.T) {
	fn := filepath.Join(GetTestingDir(), "tmp.txt")

	file, err := MakeTmpTxtFile(fn, RandInt(1000))
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}

	notify := make(chan fsnotify.Event)
	monitor := NewMonitor(GetTestingDir())

	// start monitoring thread & wait for a few seconds
	log.Print("starting monitoring thread...")
	go func() {
		monitor.WatchDrive(notify, GetTestingDir())
	}()
	time.Sleep(time.Second * 2)

	// alter the file to generate a detection
	log.Print("altering test file...")
	if err := file.Save([]byte(txtData)); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// verify detection occurred
	event := <-notify
	if !event.Has(fsnotify.Write) {
		Fail(t, GetTestingDir(), fmt.Errorf("failed to detect file write"))
	}

	// // shut down the monitoring thread
	// shutDown <- true

	// clean up
	if err := Clean(t, GetTestingDir()); err != nil {
		log.Fatal(err)
	}
}
