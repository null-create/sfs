package server

import (
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
)

func TestServerStartUpWithAPing(t *testing.T) {
	BuildEnv(true)

	// shut down signal to the server
	sig := make(chan bool)

	// start server in its own goroutine
	testServer := NewServer()
	go func() {
		testServer.TestRun(sig)
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
		sig <- true // shut down test server
		t.Fatal(err)
	}
	assert.Equal(t, res.StatusCode, http.StatusOK)

	// TODO: figure out how to send an os signal via a channel to the server
	// to shut it down
	sig <- true // shut down test server
}
