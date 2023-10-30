package client

import (
	"log"
	"testing"
	"time"

	"github.com/sfs/pkg/env"
	"github.com/sfs/pkg/monitor"
)

func TestStartHandler(t *testing.T) {
	env.BuildEnv(false)

	// make sure we clean the right testing directory
	e := env.NewE()
	tmpDir, err := e.Get("CLIENT_ROOT")
	if err != nil {
		t.Fatal(err)
	}

	// create a temp client with test files and subdirectories
	c, err := Init(true)
	if err != nil {
		Fail(t, tmpDir, err)
	}
	c.Drive.Root = MakeTmpDirs(t)

	// randomly pick a file to monitor
	files, err := c.Drive.Root.GetFiles()
	if err != nil {
		if err2 := Clean(t, GetTestingDir()); err2 != nil {
			log.Printf("[ERROR] %v", err)
			log.Fatal(err2)
		}
		Fail(t, tmpDir, err)
	}
	f := files[RandInt(len(files)-1)]

	// this is just so the event handler can get the fileID
	if err := c.Db.AddFile(f); err != nil {
		if err2 := Clean(t, GetTestingDir()); err2 != nil {
			log.Fatal(err2)
		}
		Fail(t, tmpDir, err)
	}

	// create a new monitor and watch for changes
	c.Monitor = monitor.NewMonitor(c.Drive.Root.Path)
	c.Monitor.WatchFile(f.Path)

	// create a new handler and start listening for file events from the monitor
	if err := c.NewHandler(f.Path); err != nil {
		if err2 := Clean(t, GetTestingDir()); err2 != nil {
			log.Printf("[ERROR] %v", err)
			log.Fatal(err2)
		}
		Fail(t, tmpDir, err)
	}
	if err = c.StartHandler(f.Path); err != nil {
		if err2 := Clean(t, GetTestingDir()); err2 != nil {
			log.Printf("[ERROR] %v", err)
			log.Fatal(err2)
		}
		Fail(t, tmpDir, err)
	}

	// wait a couple seconds for things to sync up
	time.Sleep(2 * time.Second)

	// modify the file to generate detections
	MutateFile(t, f)

	// wait for the monitor to register detections
	// and send them to the listener
	time.Sleep(2 * time.Second)

	if err := Clean(t, tmpDir); err != nil {
		if err2 := Clean(t, GetTestingDir()); err2 != nil {
			log.Printf("[ERROR] %v", err)
			log.Fatal(err2)
		}
		// reset our .env file for other tests
		if err2 := e.Set("CLIENT_NEW_SERVICE", "true"); err2 != nil {
			log.Fatal(err2)
		}
		log.Fatal(err)
	}
}

// func TestBuildHandlers(t *testing.T) {}
