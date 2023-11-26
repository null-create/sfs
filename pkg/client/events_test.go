package client

import (
	"log"
	"testing"
	"time"

	"github.com/sfs/pkg/env"
)

func TestStartHandler(t *testing.T) {
	env.BuildEnv(false)

	// make sure we clean the right testing directory
	e := env.NewE()
	tmpDir, err := e.Get("CLIENT_ROOT")
	if err != nil {
		t.Fatal(err)
	}

	// create a temp client with test files and subdirectories,
	// and build a sync index since the event handler will need
	// to interact with it
	c, err := Init(true)
	if err != nil {
		Fail(t, tmpDir, err)
	}
	c.Drive.Root = MakeTmpDirs(t)
	if err := c.Drive.BuildSyncIdx(); err != nil {
		Fail(t, tmpDir, err)
	}

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
	c.Db.WhichDB("files")
	if err := c.Db.AddFile(f); err != nil {
		if err2 := Clean(t, GetTestingDir()); err2 != nil {
			log.Fatal(err2)
		}
		Fail(t, tmpDir, err)
	}

	// monitor this new file for changes
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
	if err = Clean(t, GetTestingDir()); err != nil {
		log.Fatal(err)
	}
}

// func TestBuildHandlers(t *testing.T) {}
