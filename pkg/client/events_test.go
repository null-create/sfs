package client

import (
	"fmt"
	"testing"
	"time"

	"github.com/sfs/pkg/env"
)

func TestStartHandler(t *testing.T) {
	env.SetEnv(false)
	tmpDir, err := svcCfgs.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}

	client := newTestClient(t, tmpDir)

	// add a bunch of test directories to monitor
	client.Drive.Root = MakeTmpDirsWithPath(t, tmpDir, client.Drive.ID)

	// create initial sync index
	client.Drive.BuildSyncIdx()

	// randomly pick a file to monitor
	files := client.Drive.Root.GetFiles()
	if files == nil {
		Fail(t, tmpDir, fmt.Errorf("no files found"))
	}
	file := files[RandInt(len(files)-1)]

	// this is just so the event handler can get the fileID
	if err := client.Db.AddFile(file); err != nil {
		Fail(t, tmpDir, err)
	}

	// monitor this new file for changes
	if err := client.Monitor.Watch(file.Path); err != nil {
		Fail(t, tmpDir, err)
	}

	// create a new handler and start listening for file events from the monitor
	if err := client.NewHandler(file.Path); err != nil {
		Fail(t, tmpDir, err)
	}
	if err = client.StartHandler(file.Path); err != nil {
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
		if err2 := svcCfgs.Set("CLIENT_NEW_SERVICE", "true"); err2 != nil {
			t.Fatal(err2)
		}
		t.Fatal(err)
	}
	if err = Clean(t, GetTestingDir()); err != nil {
		t.Fatal(err)
	}
}
