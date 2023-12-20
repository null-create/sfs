package server

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"path/filepath"
	"testing"

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
	client := new(http.Client)
	res, err := client.Get(fmt.Sprint(LocalHost, "/v1/files/all"))
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

	// display response/results
	log.Printf("[TEST] response code: %d", res.StatusCode)
	b, err := httputil.DumpResponse(res, true)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}
	log.Printf("[TEST] response: %v", string(b))

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

	// ---- set up test service

	// so we can add the test file directly to the db ahead of time
	testSvc, err := Init(false, false)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// create tmp test drive. we'll need this
	// since the service requires a drive instance with
	// root to be found in the database in order to retrieve it
	tmpDrive := MakeTmpDriveWithPath(t, testSvc.UserDir)
	if err := testSvc.AddDrive(tmpDrive); err != nil {
		Fail(t, testSvc.UserDir, err)
	}

	// pick a file from the tmp drive to download
	files := tmpDrive.Root.GetFiles()
	if len(files) == 0 {
		Fail(t, testSvc.UserDir, fmt.Errorf("no test files found"))
	}
	file := files[RandInt(len(files)-1)]

	// ---- start server

	// shut down signal to the server
	shutDown := make(chan bool)

	// start testing server
	log.Print("starting test server...")
	testServer := NewServer()
	go func() {
		testServer.TestRun(shutDown)
	}()

	// ---- atttempt to retrieve file via its API endpoint

	log.Print("[TEST] attempting to retrieve file via its API endpoint...")
	client := new(http.Client)
	res, err := client.Get(file.Endpoint)
	if err != nil {
		shutDown <- true
		Fail(t, testSvc.UserDir, err)
	}
	if res.StatusCode != http.StatusOK {
		shutDown <- true
		b, err := httputil.DumpResponse(res, true)
		if err != nil {
			Fail(t, testSvc.UserDir, err)
		}
		msg := fmt.Sprintf(
			"response code was not 200: %d\n response: %v\n",
			res.StatusCode, string(b),
		)
		Fail(t, testSvc.UserDir, fmt.Errorf(msg))
	}

	// get file info from response body and display
	log.Printf("[TEST] response code: %d", res.StatusCode)
	b, err := httputil.DumpResponse(res, true)
	if err != nil {
		log.Printf("[TEST] failed to dump response body: %v", err)
	} else {
		log.Printf("[TEST] response: %s", string(b))
	}

	// TODO: download file and compare contents against original

	shutDown <- true // shut down test server

	// remove tmp file
	if err := Clean(testSvc.UserDir); err != nil {
		log.Fatal(err)
	}
}

func TestFileDeleteAPI(t *testing.T) {
	BuildEnv(true)

	// ---- set up test service

	// so we can add the test file directly to the db ahead of time
	testSvc, err := Init(false, false)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// create tmp test drive. we'll need this
	// since the service requires a drive instance with
	// root to be found in the database in order to retrieve it
	tmpDrive := MakeTmpDriveWithPath(t, testSvc.UserDir)
	if err := testSvc.AddDrive(tmpDrive); err != nil {
		Fail(t, testSvc.UserDir, err)
	}

	// pick a file from the tmp drive to download
	files := tmpDrive.Root.GetFiles()
	if len(files) == 0 {
		Fail(t, testSvc.UserDir, fmt.Errorf("no test files found"))
	}
	// file := files[RandInt(len(files)-1)]

	// ---- start server

	// shut down signal to the server
	shutDown := make(chan bool)

	// start testing server
	log.Print("starting test server...")
	testServer := NewServer()
	go func() {
		testServer.TestRun(shutDown)
	}()

	shutDown <- true // shut down test server

	if err := Clean(testSvc.UserDir); err != nil {
		log.Fatal(err)
	}
}

// func TestGetDirectoryAPI(t *testing.T) {}

// func TestPutDirectoryAPI(t *testing.T) {}
