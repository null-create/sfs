package client

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/sfs/pkg/env"
	svr "github.com/sfs/pkg/server"
)

func TestGetServerSyncIndex(t *testing.T) {
	// build env and get our temp working directory
	env.BuildEnv(true)
	e := env.NewE()
	tmpDir, err := e.Get("CLIENT_ROOT")
	if err != nil {
		t.Fatal(err)
	}

	// shut down signal to the test server
	shutDown := make(chan bool)

	// fire up a test server with test files
	testServer := svr.NewServer()
	go func() {
		testServer.TestRun(shutDown)
	}()

	// wait for server to be ready
	time.Sleep(2 * time.Second)

	// create tmp client
	client, err := Init(true)
	if err != nil {
		shutDown <- true
		Fail(t, tmpDir, err)
	}

	// retrieve index from server API and confirm non-empty fields
	idx := client.GetServerIdx()
	if idx == nil {
		shutDown <- true
		Fail(t, tmpDir, fmt.Errorf("server index is nil"))
	}

	// TODO: other tests...

	// shutdown test server
	shutDown <- true

	if err := Clean(t, tmpDir); err != nil {
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
