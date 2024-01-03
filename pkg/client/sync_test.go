package client

import (
	"fmt"
	"log"
	"testing"

	"github.com/sfs/pkg/env"
	svr "github.com/sfs/pkg/server"
)

func TestGetServerSyncIndex(t *testing.T) {
	// build env and get our temp working directory
	env.SetEnv(false)

	e := env.NewE()
	clientRoot, err := e.Get("CLIENT_ROOT")
	if err != nil {
		t.Fatal(err)
	}

	// create a tmp service with drive, then
	// generate a new sync index to be retrieved by the client
	tmpSvc, err := svr.Init(false, false)
	if err != nil {
		t.Fatal(err)
	}

	// create tmp client & add drive to service before contacting server
	client, err := Init(true)
	if err != nil {
		Fail(t, clientRoot, err)
	}
	if err := tmpSvc.AddDrive(client.Drive); err != nil {
		Fail(t, clientRoot, err)
	}

	// shut down signal to the test server
	shutDown := make(chan bool)

	// fire up a test server with test files
	testServer := svr.NewServer()
	go func() {
		testServer.TestRun(shutDown)
	}()
	// retrieve index from server API and confirm non-empty fields
	idx := client.GetServerIdx()
	if idx == nil {
		shutDown <- true
		if err := Clean(t, GetTestingDir()); err != nil {
			t.Fatal(err)
		}
		Fail(t, clientRoot, fmt.Errorf("failed to retrieve sync index from server"))
	}

	// TODO: other tests...

	// shutdown test server
	shutDown <- true

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Fatal(err)
	}
	if err := Clean(t, clientRoot); err != nil {
		// reset our .env file for other tests
		if err2 := e.Set("CLIENT_NEW_SERVICE", "true"); err2 != nil {
			log.Fatal(err2)
		}
		log.Fatal(err)
	}
}

func TestPush(t *testing.T) {
	// fire up a test server with *client-side* test files

	// create a sync index

	// modify tmp files

	// populate ToUpdate map

	// create a file queue with files to be sent to the server,
	// then call Push()
}

func TestPull(t *testing.T) {
	// fire up a test server with *server-side* test files

	// create a server-side sync index

	// modify tmp files

	// populate ToUpdate map

	// create a tmp client

	// retrieve index from server API and confirm non-empty fields
}
