package client

import (
	"fmt"
	"testing"
	"time"

	"github.com/sfs/pkg/env"
)

func TestStartHandler(t *testing.T) {
	env.SetEnv(false)

	// make sure we clean the right testing directory
	e := env.NewE()
	tmpDir, err := e.Get("CLIENT_ROOT")
	if err != nil {
		t.Fatal(err)
	}

	// create a temp client with test files and subdirectories,
	// and build a sync index since the event handler will need
	// to interact with it
	client, err := Init(true)
	if err != nil {
		Fail(t, tmpDir, err)
	}
	client.Drive.Root = MakeTmpDirs(t)

	// create initial sync index
	client.Drive.BuildSyncIdx()

	// randomly pick a file to monitor
	files := client.Drive.Root.GetFiles()
	if files == nil {
		if err2 := Clean(t, GetTestingDir()); err2 != nil {
			t.Fatal(err2)
		}
		Fail(t, tmpDir, fmt.Errorf("no files found"))
	}
	file := files[RandInt(len(files)-1)]

	// this is just so the event handler can get the fileID
	if err := client.Db.AddFile(file); err != nil {
		if err2 := Clean(t, GetTestingDir()); err2 != nil {
			t.Fatal(err2)
		}
		Fail(t, tmpDir, err)
	}

	// monitor this new file for changes
	if err := client.Monitor.WatchItem(file.Path); err != nil {
		if err2 := Clean(t, GetTestingDir()); err2 != nil {
			t.Fatal(err2)
		}
		Fail(t, tmpDir, err)
	}

	// create a new handler and start listening for file events from the monitor
	if err := client.NewHandler(file.Path); err != nil {
		if err2 := Clean(t, GetTestingDir()); err2 != nil {
			t.Fatal(err2)
		}
		Fail(t, tmpDir, err)
	}
	if err = client.StartHandler(file.Path); err != nil {
		if err2 := Clean(t, GetTestingDir()); err2 != nil {
			t.Fatal(err2)
		}
		Fail(t, tmpDir, err)
	}

	// wait a couple seconds for things to sync up
	time.Sleep(2 * time.Second)

	// modify the file to generate detections
	MutateFile(t, file)

	// wait for the monitor to register detections
	// and send them to the listener
	time.Sleep(2 * time.Second)

	if err := Clean(t, tmpDir); err != nil {
		if err2 := Clean(t, GetTestingDir()); err2 != nil {
			t.Fatal(err2)
		}
		// reset our .env file for other tests
		if err2 := e.Set("CLIENT_NEW_SERVICE", "true"); err2 != nil {
			t.Fatal(err2)
		}
		t.Fatal(err)
	}
	if err = Clean(t, GetTestingDir()); err != nil {
		t.Fatal(err)
	}
}
