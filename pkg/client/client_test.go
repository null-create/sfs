package client

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/sfs/pkg/env"

	"github.com/alecthomas/assert/v2"
)

func TestNewClient(t *testing.T) {
	env.BuildEnv(true)

	// make sure we clean the right testing directory
	e := env.NewE()
	tmpDir, err := e.Get("CLIENT_ROOT")
	if err != nil {
		t.Fatal(err)
	}

	// initialize a new client
	client, err := Init(true)
	if err != nil {
		Fail(t, tmpDir, err)
	}
	assert.NotEqual(t, nil, client)
	assert.NotEqual(t, nil, client.Conf)
	assert.NotEqual(t, "", client.User)
	assert.NotEqual(t, "", client.UserID)
	assert.NotEqual(t, "", client.SfDir)
	assert.NotEqual(t, nil, client.Db)
	assert.NotEqual(t, nil, client.client)

	// check that .env was updated after initialization,
	// specifically that CLIENT_NEW_SERVICE was set to "false"
	v, err := e.Get("CLIENT_NEW_SERVICE")
	if err != nil {
		Fail(t, tmpDir, err)
	}
	assert.Equal(t, "false", v)

	// check for service directories and necessary databases
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		Fatal(t, err)
	}
	assert.NotEqual(t, 0, len(entries))

	if err := Clean(t, tmpDir); err != nil {
		// reset our .env file for other tests
		if err2 := e.Set("CLIENT_NEW_SERVICE", "true"); err2 != nil {
			log.Fatal(err2)
		}
		log.Fatal(err)
	}
}

func TestLoadClient(t *testing.T) {
	env.BuildEnv(true)

	e := env.NewE()
	tmpDir, err := e.Get("CLIENT_ROOT")
	if err != nil {
		t.Fatal(err)
	}

	// initialize a new client, then load a new client object
	// from the initialized service directory and state file
	c1, err := Init(true)
	if err != nil {
		Fail(t, tmpDir, err)
	}

	c2, err := Init(false)
	if err != nil {
		Fail(t, tmpDir, err)
	}
	assert.NotEqual(t, nil, c2)
	assert.Equal(t, c1.Conf, c2.Conf)
	assert.Equal(t, c1.User, c2.User)
	assert.Equal(t, c1.UserID, c2.UserID)
	assert.Equal(t, c1.SfDir, c2.SfDir)
	// assert.Equal(t, c1.Db, c2.Db)
	// assert.Equal(t, c1.client, c2.client)

	if err := Clean(t, tmpDir); err != nil {
		// reset our .env file for other tests
		if err2 := e.Set("CLIENT_NEW_SERVICE", "true"); err2 != nil {
			log.Fatal(err2)
		}
		log.Fatal(err)
	}
}

func TestLoadClientSaveState(t *testing.T) {
	env.BuildEnv(true)

	// make sure we clean the right testing directory
	e := env.NewE()
	tmpDir, err := e.Get("CLIENT_ROOT")
	if err != nil {
		t.Fatal(err)
	}

	// initialize a new client, then load a new client object
	// from the initialized service directory and state file
	tmpClient, err := Init(true)
	if err != nil {
		Fail(t, tmpDir, err)
	}
	if err := tmpClient.SaveState(); err != nil {
		Fail(t, tmpDir, err)
	}
	entries, err := os.ReadDir(tmpClient.SfDir)
	if err != nil {
		Fail(t, tmpDir, err)
	}
	assert.NotEqual(t, 0, len(entries))
	assert.Equal(t, 1, len(entries)) // should only have 1 state file at a time
	assert.True(t, strings.Contains(entries[0].Name(), "client-state"))
	assert.True(t, strings.Contains(entries[0].Name(), ".json"))

	if err := Clean(t, tmpDir); err != nil {
		// reset our .env file for other tests
		if err2 := e.Set("CLIENT_NEW_SERVICE", "true"); err2 != nil {
			log.Fatal(err2)
		}
		log.Fatal(err)
	}
}

func TestClientAddNewUser(t *testing.T) {}

func TestClientUpdateUser(t *testing.T) {}

func TestClientDeleteUser(t *testing.T) {}

func TestClientBuildSyncIndex(t *testing.T) {}

func TestClientBuildAndUpdateSyncIndex(t *testing.T) {}

func TestClientDirectoryMonitor(t *testing.T) {}

func TestClientContactServer(t *testing.T) {}

func TestClientGetSyncIndexFromServer(t *testing.T) {}

func TestClientSendSyncIndexToServer(t *testing.T) {}
