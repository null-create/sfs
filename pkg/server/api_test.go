package server

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"path/filepath"
	"testing"
	"time"

	"github.com/sfs/pkg/transfer"
)

const LocalHost = "http://localhost:8080"

func TestGetAllFileInfoAPI(t *testing.T) {
	BuildEnv(true)

	// shut down signal to the server
	shutDown := make(chan bool)

	// start testing server
	log.Print("[TEST] starting test server...")
	testServer := NewServer()
	go func() {
		testServer.TestRun(shutDown)
	}()

	// attempt to retrieve all file info from the server
	log.Printf("[TEST] retrieving file data...")

	endpoint := fmt.Sprint(LocalHost, "/v1/files/all")

	client := http.Client{Timeout: time.Second * 600}
	res, err := client.Get(endpoint)
	if err != nil {
		shutDown <- true
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		shutDown <- true
		b, err := httputil.DumpResponse(res, true)
		if err != nil {
			t.Fatal(err)
		}
		msg := fmt.Sprintf(
			"response code was not 200: %d\n response: %v\n",
			res.StatusCode, string(b),
		)
		t.Fatal(fmt.Errorf(msg))
	}

	// retrieve data from the request
	b, err := httputil.DumpResponse(res, true)
	if err != nil {
		shutDown <- true
		t.Fatal(err)
	}
	log.Printf("[TEST] retrieved file data: %v", string(b))

	log.Print("[TEST] shutting down test server...")
	shutDown <- true
}

func TestFileGetAPI(t *testing.T) {
	BuildEnv(true)

	// shut down signal to the server
	shutDown := make(chan bool)

	// add temp file to try and retrieve

	// start testing server
	log.Print("starting test server...")
	testServer := NewServer()
	go func() {
		testServer.TestRun(shutDown)
	}()

	shutDown <- true // shut down test server

	// remove tmp file
}

func TestFilePutAPI(t *testing.T) {
	BuildEnv(true)

	// shut down signal to the server
	shutDown := make(chan bool)

	// start testing server
	log.Print("[TEST] starting test server...")
	testServer := NewServer()
	go func() {
		testServer.TestRun(shutDown)
	}()

	// create tmp file to try and send it to the server
	log.Print("[TEST] creating tmp file...")
	file, err := MakeTmpTxtFile(filepath.Join(GetTestingDir(), "tmp.txt"), RandInt(1000))
	if err != nil {
		shutDown <- true
		Fail(t, GetTestingDir(), err)
	}

	endpoint := fmt.Sprint(LocalHost, "/v1/files/new")

	// transfer file
	log.Print("[TEST] uploading file...")
	transfer := transfer.NewTransfer()
	if err := transfer.Upload(http.MethodPut, file, endpoint); err != nil {
		shutDown <- true // shut down test server
		Fail(t, GetTestingDir(), err)
	}

	log.Print("[TEST] retrieving info about file from server...")

	// confirm file's presence via a GET
	fileEndpoint := fmt.Sprint(LocalHost, fmt.Sprintf("/v1/files/%s", file.ID))

	client := http.Client{Timeout: time.Second * 600}
	res, err := client.Get(fileEndpoint)
	if err != nil {
		shutDown <- true
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		shutDown <- true
		msg := fmt.Sprintf(
			"response code was not 200: %d\n header object: %v\n",
			res.StatusCode, res.Header,
		)
		t.Fatal(fmt.Errorf(msg))
	}

	shutDown <- true // shut down test server

	// clean up
	if err := Clean(GetTestingDir()); err != nil {
		log.Fatal(err)
	}
}

func TestFileDeleteAPI(t *testing.T) {
	BuildEnv(true)

	// shut down signal to the server
	shutDown := make(chan bool)

	// start testing server
	log.Print("[TEST] starting test server...")
	testServer := NewServer()
	go func() {
		testServer.TestRun(shutDown)
	}()

	// add some test files so we can retrieve one of them
	tmp := MakeTmpDirs(t)

	shutDown <- true // shut down test server

	if err := Clean(tmp.Path); err != nil {
		log.Fatal(err)
	}
}

// func TestGetDirectoryAPI(t *testing.T) {}

// func TestPutDirectoryAPI(t *testing.T) {}
