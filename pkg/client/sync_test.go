package client

import "testing"

func TestGetServerSyncIndex(t *testing.T) {
	// fire up a test server with test files

	// server creates initial sync index

	// create tmp client

	// retrieve index from server API and confirm non-empty fields
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
