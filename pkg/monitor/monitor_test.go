package monitor

import (
	"log"
	"path/filepath"
	"testing"
)

func TestMonitorWithOneFile(t *testing.T) {
	fn := filepath.Join(GetTestingDir(), "tmp.txt")

	file, err := MakeTmpTxtFile(fn, RandInt(1000))
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// shutDown := make(chan bool)
	// monitor := NewMonitor(GetTestingDir())

	// start monitoring thread & wait for a few seconds
	log.Print("starting monitoring thread...")

	// alter the file to generate a detection
	log.Print("altering test file...")
	if err := file.Save([]byte(txtData)); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// verify detection occurred

	// shut down the monitoring thread
	// shutDown <- true
	// close(shutDown)

	// clean up
	if err := Clean(t, GetTestingDir()); err != nil {
		log.Fatal(err)
	}
}

func TestMonitorWithManyFiles(t *testing.T) {}
