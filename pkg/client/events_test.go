package client

import (
	"log"
	"testing"
	"time"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/env"
	"github.com/sfs/pkg/monitor"
)

func TestStartHandler(t *testing.T) {
	env.BuildEnv(true)

	// make sure we clean the right testing directory
	e := env.NewE()
	tmpDir, err := e.Get("CLIENT_ROOT")
	if err != nil {
		t.Fatal(err)
	}

	// temp client with temp root
	user, err := e.Get("CLIENT")
	if err != nil {
		t.Fatal(err)
	}

	// create a temp client (no actual service directories)
	// with test files and subdirectories
	c := NewClient(user, auth.NewUUID())
	c.Drive.Root = MakeTmpDirs(t)

	// randomly pick a file to monitor
	files, err := c.Drive.Root.GetFiles()
	if err != nil {
		Fail(t, tmpDir, err)
	}
	f := files[RandInt(len(files)-1)]

	// create a new monitor and watch for changes
	c.Monitor = monitor.NewMonitor(c.Drive.Root.Path)
	c.Monitor.WatchFile(f.Path)

	// create a new handler and start listening for file events from the monitor
	if err := c.NewHandler(f.Path); err != nil {
		Fail(t, tmpDir, err)
	}
	if err = c.StartHandler(f.Path); err != nil {
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
		// reset our .env file for other tests
		if err2 := e.Set("CLIENT_NEW_SERVICE", "true"); err2 != nil {
			log.Fatal(err2)
		}
		log.Fatal(err)
	}
}

func TestBuildHandlers(t *testing.T) {}
