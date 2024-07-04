package client

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/env"
	"github.com/sfs/pkg/server"
	svc "github.com/sfs/pkg/service"

	"github.com/alecthomas/assert/v2"
)

var e = env.NewE()

func newTestClient(t *testing.T) *Client {
	tmpDir, err := e.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}
	tmpClient, err := SetupClient(tmpDir)
	if err != nil {
		Fail(t, tmpDir, err)
	}
	return tmpClient
}

// create a new client without a user
func TestNewClient(t *testing.T) {
	env.SetEnv(false)

	client := newTestClient(t)

	tmpDir, err := e.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}

	assert.NotEqual(t, nil, client)
	assert.NotEqual(t, nil, client.Conf)
	assert.NotEqual(t, nil, client.User)
	assert.NotEqual(t, "", client.Root)
	assert.NotEqual(t, "", client.SfDir)
	assert.NotEqual(t, nil, client.Monitor)
	assert.NotEqual(t, nil, client.Drive)
	assert.NotEqual(t, nil, client.Db)
	assert.NotEqual(t, nil, client.Handlers)
	assert.NotEqual(t, nil, client.Transfer)

	// check that .env was updated after initialization,
	// specifically that CLIENT_NEW_SERVICE was set to "false"
	v, err := e.Get("CLIENT_NEW_SERVICE")
	if err != nil {
		Fail(t, tmpDir, err)
	}
	assert.Equal(t, "false", v)

	// check for service directories and necessary databases
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		Fatal(t, err)
	}
	assert.NotEqual(t, 0, len(entries))

	if err := Clean(t, tmpDir); err != nil {
		log.Fatal(err)
	}
}

// load a client with a pre-existing user
func TestLoadClient(t *testing.T) {
	env.SetEnv(false)

	tmpDir, err := e.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}

	// initialize a new client, then load a new client object
	// from the initialized service directory and state file
	c1 := newTestClient(t)

	// add a new user
	newUser, err := newUser()
	if err != nil {
		Fail(t, tmpDir, err)
	}
	newUser.DriveID = c1.Drive.ID
	c1.User = newUser
	if err = c1.SaveState(); err != nil {
		Fail(t, tmpDir, err)
	}

	// start a new client with this data and compare
	c2, err := Init(false)
	if err != nil {
		Fail(t, tmpDir, err)
	}

	// clean up before any possible assert failures
	if err := Clean(t, tmpDir); err != nil {
		log.Fatal(err)
	}

	assert.NotEqual(t, nil, c2)
	assert.Equal(t, c1.Conf, c2.Conf)
	assert.Equal(t, c1.User, c2.User)
	assert.Equal(t, c1.Root, c2.Root)
	assert.Equal(t, c1.User.ID, c2.User.ID)
	assert.Equal(t, c1.SfDir, c2.SfDir)
	assert.NotEqual(t, nil, c2.Db)
	assert.True(t, c2.Db.Singleton)
}

func TestLoadClientSaveState(t *testing.T) {
	env.SetEnv(false)

	tmpDir, err := e.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}
	// initialize a new client, then load a new client object
	// from the initialized service directory and state file
	tmpClient := newTestClient(t)
	if err := tmpClient.SaveState(); err != nil {
		Fail(t, tmpDir, err)
	}

	entries, err := os.ReadDir(tmpClient.SfDir)
	if err != nil {
		Fail(t, tmpDir, err)
	}

	// Pre-emptively clean in case there's any assert failures.
	if err := Clean(t, tmpDir); err != nil {
		log.Fatal(err)
	}

	assert.NotEqual(t, 0, len(entries))
	assert.Equal(t, 1, len(entries)) // should only have 1 state file at a time
	assert.True(t, strings.Contains(entries[0].Name(), "client-state"))
	assert.True(t, strings.Contains(entries[0].Name(), ".json"))
}

func TestLoadAndStartClient(t *testing.T) {
	env.SetEnv(false)

	// use testing directory for test services
	tmpDir, err := e.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}

	// initialize and load client
	// initialize a new client, then load a new client object
	// from the initialized service directory and state file
	tmpClient := newTestClient(t)

	// start client
	go func() {
		tmpClient.Start()
	}()

	log.Print("client started...")
	time.Sleep(time.Second * 2)

	tmpClient.ShutDown()

	// clean up
	if err := Clean(t, tmpDir); err != nil {
		log.Fatal(err)
	}
}

// func TestClientAddAndRemoveFile(t *testing.T) {
// 	env.SetEnv(false)

// 	// make sure we clean the right testing directory
// 	tmpDir, err := e.Get("CLIENT_ROOT")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	tmpClient, err := Init(false)
// 	if err != nil {
// 		Fail(t, tmpDir, err)
// 	}

// 	// run a test server to register the file when added to the client
// 	shutDown := make(chan bool)
// 	testServer := server.NewServer()
// 	go func() {
// 		testServer.Start(shutDown)
// 	}()

// 	// Add a new file using the client instance
// 	testFilePath := filepath.Join(tmpDir, "test.txt")
// 	f, err := os.Create(testFilePath)
// 	if err != nil {
// 		Fail(t, tmpDir, err)
// 	}
// 	_, err = f.Write([]byte("test data"))
// 	if err != nil {
// 		Fail(t, tmpDir, err)
// 	}

// 	if err = tmpClient.AddFile(testFilePath); err != nil {
// 		Fail(t, tmpDir, err)
// 	}
// 	testFile, err := tmpClient.GetFileByPath(testFilePath)
// 	if err != nil {
// 		Fail(t, tmpDir, err)
// 	}

// 	// Remove test file
// 	if err = tmpClient.RemoveFile(testFile); err != nil {
// 		Fail(t, tmpDir, err)
// 	}

// 	if err := Clean(t, tmpDir); err != nil {
// 		// reset our .env file for other tests
// 		if err2 := e.Set("CLIENT_NEW_SERVICE", "true"); err2 != nil {
// 			log.Fatal(err2)
// 		}
// 		log.Fatal(err)
// 	}
// }

func TestClientUpdateUser(t *testing.T) {
	env.SetEnv(false)

	tmpDir, err := e.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}
	// initialize a new client, then load a new client object
	// from the initialized service directory and state file
	tmpClient := newTestClient(t)
	if err := tmpClient.SaveState(); err != nil {
		Fail(t, tmpDir, err)
	}

	// update name initialized with client
	user := tmpClient.User
	user.Name = "William J Buttlicker"

	if err := tmpClient.UpdateUser(user); err != nil {
		Fail(t, tmpDir, err)
	}

	if err := Clean(t, tmpDir); err != nil {
		log.Fatal(err)
	}
}

func TestClientDeleteUser(t *testing.T) {
	env.SetEnv(false)

	// make sure we clean the right testing directory
	tmpDir, err := e.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}
	// initialize a new client
	tmpClient := newTestClient(t)
	if err := tmpClient.SaveState(); err != nil {
		Fail(t, tmpDir, err)
	}

	user := tmpClient.User

	if err := tmpClient.RemoveUser(user.ID); err != nil {
		Fail(t, tmpDir, err)
	}

	if err := Clean(t, tmpDir); err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, nil, tmpClient.User)
}

func TestClientBuildSyncIndex(t *testing.T) {
	env.SetEnv(false)

	// make sure we clean the right testing directory
	tmpDir, err := e.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}
	// initialize a new client
	tmpClient := newTestClient(t)
	if err := tmpClient.SaveState(); err != nil {
		Fail(t, tmpDir, err)
	}

	// make a bunch of dummy files for this "client"
	total := RandInt(25)
	files := make([]*svc.File, 0, total)
	for i := 0; i < total; i++ {
		fn := filepath.Join(tmpDir, tmpClient.User.Name, "root", fmt.Sprintf("tmp-%d.txt", i))
		if file, err := MakeTmpTxtFile(fn, RandInt(1000)); err == nil {
			files = append(files, file)
		} else {
			Fail(t, tmpDir, err)
		}
	}

	// set up a new client drive and generate a last sync index of the files
	root := svc.NewDirectory("root", tmpClient.Conf.User, tmpClient.Drive.ID, tmpClient.Root)
	root.AddFiles(files)

	drv := svc.NewDrive(auth.NewUUID(), tmpClient.Conf.User, auth.NewUUID(), root.Path, root.ID, root)

	idx := drv.Root.WalkS(svc.NewSyncIndex(tmpClient.Conf.User))
	assert.NotEqual(t, nil, idx)
	assert.NotEqual(t, 0, len(idx.LastSync))
	assert.Equal(t, total, len(idx.LastSync))

	if err := Clean(t, tmpDir); err != nil {
		log.Fatal(err)
	}
}

// indexing and monitoring tests
// this test is flaky because we don't always alter
// files with each run. might need to not make it so random.
func TestClientBuildAndUpdateSyncIndex(t *testing.T) {
	env.SetEnv(false)

	// make sure we clean the right testing directory
	tmpDir, err := e.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}
	// initialize a new client
	tmpClient := newTestClient(t)
	if err := tmpClient.SaveState(); err != nil {
		Fail(t, tmpDir, err)
	}

	// make a bunch of dummy files for this "client"
	total := RandInt(25)
	files := make([]*svc.File, 0, total)
	for i := 0; i < total; i++ {
		fn := filepath.Join(tmpDir, tmpClient.User.Name, "root", fmt.Sprintf("tmp-%d.txt", i))
		if file, err := MakeTmpTxtFile(fn, RandInt(1000)); err == nil {
			files = append(files, file)
		} else {
			Fail(t, tmpClient.Root, err)
		}
	}

	// set up a new client drive and generate a last sync index of the files
	tmpClient.Drive.Root.AddFiles(files)

	// create initial sync index
	tmpClient.Drive.SyncIndex = svc.BuildRootSyncIndex(tmpClient.Drive.Root)

	// alter some files so we can mark them to be synced
	tmpClient.Drive.Root.Files = MutateFiles(t, tmpClient.Drive.Root.Files)

	// pre-emptively clean up
	if err := Clean(t, tmpDir); err != nil {
		log.Fatal(err)
	}

	// build ToUpdate map
	tmpClient.Drive.SyncIndex = svc.BuildToUpdate(tmpClient.Drive.Root, tmpClient.Drive.SyncIndex)
	assert.NotEqual(t, nil, tmpClient.Drive.SyncIndex.FilesToUpdate)
	assert.NotEqual(t, 0, len(tmpClient.Drive.SyncIndex.FilesToUpdate))
}

func TestClientRefreshDrive(t *testing.T) {
	env.SetEnv(false)

	// make sure we clean the right testing directory
	tmpDir, err := e.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}

	// initialize a new client
	tmpClient := newTestClient(t)
	if err := tmpClient.SaveState(); err != nil {
		Fail(t, tmpDir, err)
	}

	// make a bunch of dummy files for the test clinet
	tmpClient.Drive = MakeTmpDriveWithPath(t, tmpDir)

	// add some more files
	_, err = MakeABunchOfTxtFiles(RandInt(25))
	if err != nil {
		Fail(t, tmpDir, err)
	}

	// run refresh
	tmpClient.RefreshDrive()

	// clean up before asserts so nothing gets left behind if there's failures
	if err := Clean(t, GetTestingDir()); err != nil {
		log.Fatal(err)
	}
	if err := Clean(t, tmpDir); err != nil {
		log.Fatal(err)
	}
}

func TestClientDiscoverWithPath(t *testing.T) {
	env.SetEnv(false)

	// make sure we clean the right testing directory
	tmpDir, err := e.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}

	// start a test server
	shutDown := make(chan bool)
	testServer := server.NewServer()
	go func() {
		testServer.Start(shutDown)
	}()

	// initialize a new client
	tmpClient := newTestClient(t)
	if err := tmpClient.SaveState(); err != nil {
		Fail(t, tmpDir, err)
	}

	testDrive := MakeTmpDriveWithPath(t, tmpDir)
	_, err = tmpClient.DiscoverWithPath(testDrive.Root.Path)
	if err != nil {
		Fail(t, tmpDir, err)
	}

	// shut down test server
	shutDown <- true

	// clean up before asserts so nothing gets left behind if there's failures
	if err := Clean(t, GetTestingDir()); err != nil {
		log.Fatal(err)
	}
}

func TestClientSendsANewFileToTheServer(t *testing.T) {
	env.SetEnv(false)

	tmpDir, err := e.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}

	// test server
	shutDown := make(chan bool)
	testServer := server.NewServer()
	go func() {
		testServer.Start(shutDown)
	}()

	// make a test client
	tmpClient := newTestClient(t)

	// test file
	file, err := MakeTmpTxtFile(
		filepath.Join(tmpDir, tmpClient.User.Name, "root", fmt.Sprintf("%s.txt", randString(21))), RandInt(1000),
	)
	if err != nil {
		shutDown <- true
		Fail(t, tmpDir, err)
	}

	// attempt to add a file via the top level API.
	// this adds the file to the db, and attempts to register it with the server.
	if err := tmpClient.AddFile(file.Path); err != nil {
		shutDown <- true
		Fail(t, tmpDir, err)
	}

	// shut down test server
	shutDown <- true

	// clean up
	if err := Clean(t, tmpDir); err != nil {
		log.Fatal(err)
	}
}
