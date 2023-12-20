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
		Fail(t, GetTestingDir(), err)
	}
	if res.StatusCode != http.StatusOK {
		shutDown <- true
		b, err := httputil.DumpResponse(res, true)
		if err != nil {
			Fail(t, GetTestingDir(), err)
		}
		msg := fmt.Sprintf(
			"response code was not 200: %d\n response: %v\n",
			res.StatusCode, string(b),
		)
		Fail(t, GetTestingDir(), fmt.Errorf(msg))
	}
	log.Printf("[TEST] response code: %d", res.StatusCode)

	log.Print("[TEST] shutting down test server...")
	shutDown <- true
}

func TestNewFileAPI(t *testing.T) {
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

	// transfer file
	log.Print("[TEST] uploading file...")
	transfer := transfer.NewTransfer(8080)
	if err := transfer.Upload(http.MethodPost, file, fmt.Sprint(LocalHost, "/v1/files/new")); err != nil {
		shutDown <- true
		Fail(t, GetTestingDir(), err)
	}

	// shut down test server
	shutDown <- true

	// clean up
	if err := Clean(GetTestingDir()); err != nil {
		log.Fatal(err)
	}
}

func TestFileGetAPI(t *testing.T) {
	BuildEnv(true)

	// so we can add the test file directly to the db ahead of time
	testSvc, err := Init(false, false)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// shut down signal to the server
	shutDown := make(chan bool)

	// start testing server
	log.Print("starting test server...")
	testServer := NewServer()
	go func() {
		testServer.TestRun(shutDown)
	}()

	// add temp file to try and retrieve
	log.Print("[TEST] creating tmp file...")
	file, err := MakeTmpTxtFile(filepath.Join(GetTestingDir(), "tmp.txt"), RandInt(1000))
	if err != nil {
		shutDown <- true
		Fail(t, GetTestingDir(), err)
	}
	if err := testSvc.AddFile(file.OwnerID, file.DirID, file); err != nil {
		shutDown <- true
		Fail(t, GetTestingDir(), err)
	}

	// atttempt to retrieve file via its API endpoint
	log.Print("[TEST] attempting to retrieve file via its API endpoint...")
	client := http.Client{Timeout: time.Second * 600}
	res, err := client.Get(file.Endpoint)
	if err != nil {
		shutDown <- true
		Fail(t, GetTestingDir(), err)
	}
	if res.StatusCode != http.StatusOK {
		shutDown <- true
		b, err := httputil.DumpResponse(res, true)
		if err != nil {
			Fail(t, GetTestingDir(), err)
		}
		msg := fmt.Sprintf(
			"response code was not 200: %d\n response: %v\n",
			res.StatusCode, string(b),
		)
		Fail(t, GetTestingDir(), fmt.Errorf(msg))
	}

	log.Printf("[TEST] response code: %d", res.StatusCode)

	// get file info from response body and display

	shutDown <- true // shut down test server

	// remove tmp file
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
