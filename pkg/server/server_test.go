package server

import (
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/sfs/pkg/env"
)

func TestServerStartUpWithAPing(t *testing.T) {
	env.SetEnv(false)

	// shut down signal to the server
	shutDown := make(chan bool)

	// start testing server
	testServer := NewServer()
	go func() {
		testServer.TestRun(shutDown)
	}()

	// wait for server to start up
	log.Printf("waiting for server to start up...")
	time.Sleep(time.Second * 5)

	// create a basic http client and ping the server at
	// https://localhost:8080/ping
	log.Printf("sending ping to test server...")
	client := http.Client{Timeout: time.Second * 10}

	res, err := client.Get("http://localhost:8080/ping")
	if err != nil {
		shutDown <- true
		t.Fatal(err)
	}
	assert.Equal(t, res.StatusCode, http.StatusOK)

	shutDown <- true // shut down test server
}
