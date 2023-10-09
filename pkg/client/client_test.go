package client

import (
	"log"
	"testing"

	"github.com/sfs/pkg/env"

	"github.com/alecthomas/assert/v2"
)

func TestNewClient(t *testing.T) {
	env.BuildEnv(true)

	client, err := Init(true)
	if err != nil {
		if err := Clean(t, GetTestingDir()); err != nil {
			log.Fatal(err)
		}
	}
	assert.NotEqual(t, nil, client)
	assert.NotEqual(t, nil, client.Conf)

	// check that .env was updated after initialization,
	// specifically that CLIENT_NEW_SERVICE was set to "false"

	// check for service directories and necessary databases

	if err := Clean(t, client.Conf.Root); err != nil {
		log.Fatal(err)
	}
}

func TestLoadClient(t *testing.T) {}

func TestLoadClientSaveState(t *testing.T) {}

func TestClientAddNewUser(t *testing.T) {}

func TestClientUpdateUser(t *testing.T) {}

func TestClientDeleteUser(t *testing.T) {}

func TestClientBuildSyncIndex(t *testing.T) {}

func TestClientBuildAndUpdateSyncIndex(t *testing.T) {}

func TestClientDirectoryMonitor(t *testing.T) {}

func TestClientContactServer(t *testing.T) {}

func TestClientGetSyncIndexFromServer(t *testing.T) {}

func TestClientSendSyncIndexToServer(t *testing.T) {}
