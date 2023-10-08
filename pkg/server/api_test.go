package server

import (
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
)

const LocalHost = "http://localhost:8080"

func TestFileGetAPI(t *testing.T) {
	BuildEnv(true)

	// shut down signal to the server
	shutDown := make(chan bool)

	// start testing server
	testServer := NewServer()
	go func() {
		testServer.TestRun(shutDown)
	}()

	// wait for server to start up
	log.Printf("waiting for server to start up...")
	time.Sleep(time.Second * 2)

	client := http.Client{Timeout: time.Second * 10}

	log.Printf("retrieving file data...")

	endpoint := fmt.Sprint(LocalHost, "/v1/files/all")
	res, err := client.Get(endpoint)
	if err != nil {
		shutDown <- true
		t.Fatal(err)
	}
	assert.Equal(t, res.StatusCode, http.StatusOK)

	shutDown <- true // shut down test server
}

func TestFilePutAPI(t *testing.T) {}

func TestFileDeleteAPI(t *testing.T) {}

func TestGetDirectoryAPI(t *testing.T) {}

func TestPutDirectoryAPI(t *testing.T) {}
