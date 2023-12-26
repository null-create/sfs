package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sfs/pkg/env"
	"github.com/sfs/pkg/transfer"
)

const LocalHost = "http://localhost:8080"

// NOTE: this endpoint was pulled straight from the DB and might not always
// work. may have to manually update if needed.
const ServerFile = "http://localhost:8080/v1/files/i/4e539b7b-9ed7-11ee-aef3-0a0027000014"

func TestGetAllFileInfoAPI(t *testing.T) {
	env.SetEnv(false)

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
	env.SetEnv(false)

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
	if err := transfer.Upload(
		http.MethodPost,
		file,
		fmt.Sprint(LocalHost, "/v1/files/new"),
	); err != nil {
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

func TestGetSingleFileInfoAPI(t *testing.T) {
	env.SetEnv(false)

	// shut down signal to the server
	shutDown := make(chan bool)

	// start testing server
	log.Print("[TEST] starting test server...")
	testServer := NewServer()
	go func() {
		testServer.TestRun(shutDown)
	}()

	// attempt to retrieve file info about one file from the server
	log.Printf("[TEST] retrieving test file data...")
	client := new(http.Client)
	res, err := client.Get(ServerFile)
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
		log.Printf("[TEST] failed to parse response : %v", err)
	} else {
		log.Printf("[TEST] response: %v", string(b))
	}

	log.Print("[TEST] shutting down test server...")
	shutDown <- true
}

func TestFileGetAPI(t *testing.T) {
	env.SetEnv(false)

	// ---- set up test service ---------------------------------------

	// so we can add the test file directly to the db ahead of time
	testSvc, err := Init(false, false)
	if err != nil {
		Fail(t, testSvc.UserDir, fmt.Errorf("failed to initialize test service: %v", err))
	}

	// create tmp test drive. we'll need this
	// since the service requires a drive instance with
	// root to be found in the database in order to retrieve it
	tmpDrive := MakeTmpDriveWithPath(t, testSvc.UserDir)
	if err := testSvc.AddDrive(tmpDrive); err != nil {
		Fail(t, testSvc.UserDir, fmt.Errorf("failed to create test drive: %v", err))
	}

	// pick a file from the tmp drive to download
	files := tmpDrive.Root.GetFiles()
	if len(files) == 0 {
		Fail(t, testSvc.UserDir, fmt.Errorf("no test files found"))
	}
	testFile := files[RandInt(len(files)-1)]

	// ---- start server----------------------------------------------

	// shut down signal to the server
	shutDown := make(chan bool)

	// start testing server
	log.Print("[TEST] starting test server...")
	testServer := NewServer()
	go func() {
		testServer.TestRun(shutDown)
	}()

	// ---- atttempt to retrieve file via its API endpoint -----------

	// contact the server
	log.Print("[TEST] attempting to retrieve file via its API endpoint...")
	client := &http.Client{
		Timeout: 600 * time.Second,
	}
	res, err := client.Get(testFile.Endpoint)
	if err != nil {
		shutDown <- true
		Fail(t, testSvc.UserDir, fmt.Errorf("failed to contact server: %v", err))
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

	// get file info from response body
	b, err := httputil.DumpResponse(res, false)
	if err != nil {
		log.Printf("[TEST] failed to dump response: %v", err)
	} else {
		log.Printf("[TEST] server: %s", string(b))
	}
	defer res.Body.Close()

	// download file
	log.Print("[TEST] downloading test file...")
	var buf bytes.Buffer
	_, err = io.Copy(&buf, res.Body)
	if err != nil {
		Fail(t, testSvc.UserDir, fmt.Errorf("failed to copy response body: %v", err))
	}
	tmpFile, err := os.Create(filepath.Join(GetTestingDir(), "tmp.txt"))
	if err != nil {
		Fail(t, testSvc.UserDir, fmt.Errorf("failed to create destination test file: %v", err))
	}
	_, err = tmpFile.Write(buf.Bytes())
	if err != nil {
		Fail(t, testSvc.UserDir, fmt.Errorf("failed to write data to test file: %v", err))
	}

	// ----- verify file contents --------------------------------

	tmpFileData, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		Fail(t, testSvc.UserDir, fmt.Errorf("failed to read test file data: %v", err))
	}
	if len(tmpFileData) == 0 {
		Fail(t, testSvc.UserDir, fmt.Errorf("no test file data found"))
	}
	// testFile.Load()
	// if len(tmpFileData) != len(testFile.Content) {
	// 	Fail(t, testSvc.UserDir, fmt.Errorf("different byte counts. orig = %d, new = %d", testFile.Size(), len(tmpFileData)))
	// }
	tmpFile.Close()

	// ----- clean up ---------------------------------------------

	shutDown <- true // shut down test server

	// remove tmp files
	if err := Clean(testSvc.UserDir); err != nil {
		log.Fatal(err)
	}
	if err := Clean(GetTestingDir()); err != nil {
		log.Fatal(err)
	}
}

func TestFileDeleteAPI(t *testing.T) {
	env.SetEnv(false)

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

	// pick a file from the tmp drive to delete
	files := tmpDrive.Root.GetFiles()
	if len(files) == 0 {
		Fail(t, testSvc.UserDir, fmt.Errorf("no test files found"))
	}
	testFile := files[RandInt(len(files)-1)]

	// ---- start server

	// shut down signal to the server
	shutDown := make(chan bool)

	// start testing server
	log.Print("[TEST] starting test server...")
	testServer := NewServer()
	go func() {
		testServer.TestRun(shutDown)
	}()

	// ----- start test client and attempt to delete file via its API endpoint

	log.Print("[TEST] attempting to delete file via its API endpoint...")
	client := &http.Client{
		Timeout: 600 * time.Second,
	}

	req, err := http.NewRequest(http.MethodDelete, testFile.Endpoint, nil)
	if err != nil {
		Fail(t, testSvc.UserDir, fmt.Errorf("failed to create HTTP request: %v", err))
	}
	res, err := client.Do(req)
	if err != nil {
		shutDown <- true
		Fail(t, testSvc.UserDir, fmt.Errorf("failed to contact server: %v", err))
	}
	if res.StatusCode != http.StatusOK {
		shutDown <- true
		b, err := httputil.DumpResponse(res, true)
		if err != nil {
			log.Printf("failed to dump response: %v", err)
		} else {
			log.Printf("server resonse: \n%s\n", string(b))
		}
		Fail(t, testSvc.UserDir, fmt.Errorf("non 200 response code"))
	}

	shutDown <- true // shut down test server

	if err := Clean(testSvc.UserDir); err != nil {
		log.Fatal(err)
	}
}

// func TestNewDirectoryAPI(t *testing.T) {}

// func TestGetDirectoryAPI(t *testing.T) {}

// func TestPutDirectoryAPI(t *testing.T) {}

// func TestDeleteDirectoryAPI(t *testing.T) {}
