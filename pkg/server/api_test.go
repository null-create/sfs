package server

import (
	"bytes"
	"encoding/json"
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
	svc "github.com/sfs/pkg/service"
	"github.com/sfs/pkg/transfer"
)

const LocalHost = "http://localhost:8080"

// NOTE: this endpoint was pulled straight from the DB and might not always
// work. may have to manually update if needed.
const ServerFile = "http://localhost:8080/v1/files/i/4e539b7b-9ed7-11ee-aef3-0a0027000014"

func TestGetAllFileInfoAPI(t *testing.T) {
	env.SetEnv(false)

	// ------- make temp drive to retrieve info about, add to service
	testSvc, err := Init(false, false)
	if err != nil {
		t.Fatal(err)
	}

	// add test files to server dbs
	tmpDrive := MakeTmpDrive(t)
	if err := testSvc.AddDrive(tmpDrive); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// shut down signal to the server
	shutDown := make(chan bool)

	// start testing server
	log.Print("[TEST] starting test server...")
	testServer := NewServer()
	go func() {
		testServer.Start(shutDown)
	}()

	// ------ attempt to retrieve all file info from the server
	log.Printf("[TEST] retrieving file data...")
	endpoint := fmt.Sprint(LocalHost, "/v1/i/files/all/", tmpDrive.OwnerID)

	client := new(http.Client)
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
		testServer.Start(shutDown)
	}()

	// create tmp file to try and send it to the server
	log.Print("[TEST] creating tmp file...")
	file, err := MakeTmpTxtFile(filepath.Join(GetTestingDir(), "tmp.txt"), RandInt(1000))
	if err != nil {
		shutDown <- true
		Fail(t, GetTestingDir(), err)
	}

	// TODO: use client for this instead of transfer component.
	// create empty client, add a file, then send to server.

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
		testServer.Start(shutDown)
	}()

	// attempt to retrieve file info about one file from the server
	log.Printf("[TEST] retrieving test file data...")
	client := new(http.Client)
	client.Timeout = time.Second * 600

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
		testServer.Start(shutDown)
	}()

	// ---- atttempt to retrieve file via its API endpoint -----------

	// contact the server
	log.Print("[TEST] attempting to retrieve file via its API endpoint...")
	client := new(http.Client)
	client.Timeout = time.Second * 30

	res, err := client.Get(testFile.Endpoint)
	if err != nil {
		shutDown <- true
		Fail(t, testSvc.UserDir, fmt.Errorf("failed to contact server: %v", err))
	}
	if res.StatusCode != http.StatusOK {
		shutDown <- true
		b, err := httputil.DumpResponse(res, true)
		if err != nil {
			log.Printf("[TEST] failed to dump response: %v", err)
			Fail(t, testSvc.UserDir, fmt.Errorf("response code was not 200: %v", res.StatusCode))
		} else {
			msg := fmt.Sprintf(
				"response code was not 200: %d\n response: %v\n",
				res.StatusCode, string(b),
			)
			Fail(t, testSvc.UserDir, fmt.Errorf(msg))
		}
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
		Fail(t, testSvc.UserDir, fmt.Errorf("no test file data received"))
	}

	// TODO: more germaine file content tests

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
		testServer.Start(shutDown)
	}()

	// ----- start test client and attempt to delete file via its API endpoint

	log.Print("[TEST] attempting to delete file via its API endpoint...")
	client := new(http.Client)
	client.Timeout = time.Second * 1200 // 20 min timeout lol

	var buf bytes.Buffer
	req, err := http.NewRequest(http.MethodDelete, testFile.Endpoint, &buf)
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

	// shut down test server
	shutDown <- true

	// remove tmp drive from service
	if err := testSvc.RemoveDrive(tmpDrive.ID); err != nil {
		if err2 := Clean(testSvc.UserDir); err2 != nil {
			log.Printf("[TEST] failed to clean test service files: %v", err2)
		}
	}

	// clean up
	if err := Clean(testSvc.UserDir); err != nil {
		log.Fatal(err)
	}
}

// func TestNewDirectoryAPI(t *testing.T) {}

// func TestGetDirectoryAPI(t *testing.T) {}

// func TestPutDirectoryAPI(t *testing.T) {}

// func TestDeleteDirectoryAPI(t *testing.T) {}

func TestGetServerSyncIndex(t *testing.T) {
	env.SetEnv(false)

	// ---- set up test service

	log.Print("[TEST] initializing tmp service...")

	// so we can add the test file directly to the db ahead of time
	testSvc, err := Init(false, false)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// create tmp test drive. we'll need this
	// since the service requires a drive instance with
	// root to be found in the database in order to retrieve it
	tmpDrive := MakeTmpDriveWithPath(t, testSvc.UserDir)

	// add drive generates a sync index in addition to
	// adding the drive to the service
	if err := testSvc.AddDrive(tmpDrive); err != nil {
		Fail(t, testSvc.UserDir, err)
	}

	// ---- start server

	// shut down signal to the server
	shutDown := make(chan bool)

	// start testing server
	log.Print("[TEST] starting test server...")
	testServer := NewServer()
	go func() {
		testServer.Start(shutDown)
	}()

	// ------ create a client and contact the indexing API endpoint

	client := &http.Client{Timeout: time.Second * 600} // ten minute time out
	buf := new(bytes.Buffer)
	req, err := http.NewRequest(http.MethodGet, LocalHost+"/v1/sync/"+tmpDrive.ID, buf)
	if err != nil {
		Fail(t, testSvc.UserDir, err)
	}
	resp, err := client.Do(req)
	if err != nil {
		Fail(t, testSvc.UserDir, err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("[TEST] failed to get index from server: %d", resp.StatusCode)
		b, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Printf("[TEST] failed to dump server response: %v", err)
		} else {
			log.Printf("%s", string(b))
		}
		Fail(t, testSvc.UserDir, fmt.Errorf("failed to get index from server"))
	}

	// ----- retrieve index and verify
	var idxBuf bytes.Buffer
	_, err = io.Copy(&idxBuf, resp.Body)
	if err != nil {
		Fail(t, testSvc.UserDir, err)
	}
	var idx *svc.SyncIndex
	if err := json.Unmarshal(idxBuf.Bytes(), &idx); err != nil {
		Fail(t, testSvc.UserDir, err)
	}
	log.Print("[TEST] retrieved index:")
	log.Print(idx.ToString())

	// clean up
	if err := Clean(testSvc.UserDir); err != nil {
		log.Fatal(err)
	}
}
