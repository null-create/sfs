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

func newTestClient(t *testing.T, tmpDir string) *Client {
	tmpClient, err := SetupClient(tmpDir)
	if err != nil {
		Fail(t, tmpDir, err)
	}

	// override some defaults so our test client doesn't interact
	// with the client's resources defined in .env configs
	var tmpSvcPath = filepath.Join(tmpDir, tmpClient.User.Name)

	tmpClient.Root = filepath.Join(tmpSvcPath, "root")
	tmpClient.SfDir = filepath.Join(tmpSvcPath, "state")
	tmpClient.Db.DBPath = filepath.Join(tmpSvcPath, "dbs")
	tmpClient.LocalBackupDir = filepath.Join(tmpSvcPath, "backups")
	tmpClient.RecycleBin = filepath.Join(tmpSvcPath, "recycle")
	tmpClient.Drive.Root.Path = filepath.Join(tmpSvcPath, "root")
	tmpClient.Drive.Root.ServerPath = filepath.Join(tmpSvcPath, "root")
	tmpClient.Drive.Root.ClientPath = filepath.Join(tmpSvcPath, "root")
	tmpClient.Drive.Root.BackupPath = tmpClient.LocalBackupDir
	tmpClient.Drive.RecycleBin = tmpClient.RecycleBin

	return tmpClient
}

// create a new client without a user
func TestNewClient(t *testing.T) {
	env.SetEnv(false)
	tmpDir, err := envCfgs.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}

	client := newTestClient(t, tmpDir)

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
	v, err := envCfgs.Get("CLIENT_NEW_SERVICE")
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

func TestLoadClientSaveState(t *testing.T) {
	env.SetEnv(false)
	tmpDir, err := envCfgs.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}
	// initialize a new client, then load a new client object
	// from the initialized service directory and state file
	tmpClient := newTestClient(t, tmpDir)
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
	tmpDir, err := envCfgs.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}

	// initialize and load client
	// initialize a new client, then load a new client object
	// from the initialized service directory and state file
	tmpClient := newTestClient(t, tmpDir)

	// start client
	go func() {
		if err := tmpClient.Start(); err != nil {
			Fail(t, tmpDir, err)
		}
	}()

	log.Print("client started...")
	time.Sleep(time.Second * 2)

	// shutdown
	tmpClient.ShutDown()

	// clean up
	if err := Clean(t, tmpDir); err != nil {
		log.Fatal(err)
	}
}

func TestClientUpdateUser(t *testing.T) {
	env.SetEnv(false)
	tmpDir, err := envCfgs.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}
	// initialize a new client, then load a new client object
	// from the initialized service directory and state file
	tmpClient := newTestClient(t, tmpDir)
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
	tmpDir, err := envCfgs.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}
	// initialize a new client
	tmpClient := newTestClient(t, tmpDir)
	if err := tmpClient.SaveState(); err != nil {
		Fail(t, tmpDir, err)
	}

	user := tmpClient.User

	if err := tmpClient.RemoveUser(user.ID); err != nil {
		Fail(t, tmpDir, err)
	}

	dbUser, err := tmpClient.Db.GetUser(user.ID)
	if err != nil {
		Fail(t, tmpDir, err)
	}

	if err := Clean(t, tmpDir); err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, nil, tmpClient.User)
	assert.Equal(t, nil, dbUser)
}

func TestAddFileToClient(t *testing.T) {
	env.SetEnv(false)
	tmpDir, err := envCfgs.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}

	// initialize a new testing client
	tmpClient := newTestClient(t, tmpDir)
	if err := tmpClient.SaveState(); err != nil {
		Fail(t, tmpDir, err)
	}

	// make and add test file
	testFile, err := MakeTmpTxtFile(filepath.Join(tmpClient.Root, "tmp.txt"), RandInt(1000))
	if err != nil {
		Fail(t, tmpDir, err)
	}
	if err := tmpClient.AddFile(testFile.ClientPath); err != nil {
		Fail(t, tmpDir, err)
	}

	// pre-emptive cleanup
	if err := Clean(t, tmpDir); err != nil {
		log.Fatal(err)
	}

	assert.NotEqual(t, nil, tmpClient.Drive.Root)
	assert.NotEqual(t, 0, len(tmpClient.Drive.Root.Files))
}

func TestAddAndRemoveLocalFileFromClient(t *testing.T) {
	env.SetEnv(false)
	tmpDir, err := envCfgs.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}

	// initialize a new testing client
	tmpClient := newTestClient(t, tmpDir)
	if err := tmpClient.SaveState(); err != nil {
		Fail(t, tmpDir, err)
	}

	// make and add test file
	testFile, err := MakeTmpTxtFile(filepath.Join(tmpClient.Root, "tmp.txt"), RandInt(1000))
	if err != nil {
		Fail(t, tmpDir, err)
	}
	if err := tmpClient.AddFile(testFile.ClientPath); err != nil {
		Fail(t, tmpDir, err)
	}

	// get the file from the service, since params like dirID aren't
	// set in testFile
	tf, err := tmpClient.GetFileByPath(testFile.ClientPath)
	if err != nil {
		Fail(t, tmpDir, err)
	}
	if tf == nil {
		Fail(t, tmpDir, fmt.Errorf("test file not found in service"))
	}

	// remove test file and close DB connections
	if err := tmpClient.RemoveFile(tf); err != nil {
		Fail(t, tmpDir, err)
	}
	tmpClient.ShutDown()

	// NOTE: currently failing during Clean() because the files db is still being
	// used by "another process" which I assume is the temp client,
	// but all DB connections should be closed with a call to tmpClient.ShutDown()

	// pre-emptive cleanup
	if err := Clean(t, tmpDir); err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, 0, len(tmpClient.Drive.Root.Files))
}

func TestAddItemWithAFile(t *testing.T) {
	env.SetEnv(false)
	tmpDir, err := envCfgs.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}

	// initialize a new testing client
	tmpClient := newTestClient(t, tmpDir)
	if err := tmpClient.SaveState(); err != nil {
		Fail(t, tmpDir, err)
	}

	// make and add test file
	testFile, err := MakeTmpTxtFile(filepath.Join(tmpClient.Root, "tmp.txt"), RandInt(1000))
	if err != nil {
		Fail(t, tmpDir, err)
	}
	if err := tmpClient.AddItem(testFile.Path); err != nil {
		Fail(t, tmpDir, err)
	}

	rootEntries, err := os.ReadDir(tmpClient.Root)
	if err != nil {
		Fail(t, tmpDir, err)
	}
	assert.NotEqual(t, 0, len(rootEntries))

	backupEntries, err := os.ReadDir(tmpClient.LocalBackupDir)
	if err != nil {
		Fail(t, tmpDir, err)
	}
	assert.NotEqual(t, 0, len(backupEntries))

	if err := Clean(t, tmpDir); err != nil {
		log.Fatal(err)
	}
}

func TestClientRemoveDir(t *testing.T) {
	env.SetEnv(false)
	tmpDir, err := envCfgs.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}

	// initialize a new testing client
	tmpClient := newTestClient(t, tmpDir)
	if err := tmpClient.SaveState(); err != nil {
		Fail(t, tmpDir, err)
	}

	// make a test directory with files and a subdirectory within the client
	tmpDrive := MakeTmpDriveWithPath(t, tmpClient.Drive.Root.ClientPath)
	if err := tmpClient.AddDrive(tmpDrive); err != nil {
		Fail(t, tmpDir, err)
	}

	if err := tmpClient.RemoveDir(tmpDrive.Root); err != nil {
		Fail(t, tmpDir, err)
	}

	// make sure everything was actually removed
	dbFiles, err := tmpClient.Db.GetUsersFiles(tmpClient.UserID)
	if err != nil {
		Fail(t, tmpDir, err)
	}
	assert.Equal(t, 0, len(dbFiles))

	dbDirs, err := tmpClient.Db.GetUsersDirectories(tmpClient.UserID)
	if err != nil {
		Fail(t, tmpDir, err)
	}
	assert.Equal(t, 0, len(dbDirs))

	if err := Clean(t, tmpDir); err != nil {
		log.Fatal(err)
	}
}

func TestUpdateBackupDirs(t *testing.T) {
	env.SetEnv(false)
	tmpDir, err := envCfgs.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}

	// initialize a new testing client
	tmpClient := newTestClient(t, tmpDir)
	if err := tmpClient.SaveState(); err != nil {
		Fail(t, tmpDir, err)
	}

	// make a test directory with files and a subdirectory within the client
	tmpDrive := MakeTmpDriveWithPath(t, tmpClient.Drive.Root.ClientPath)
	if err := tmpClient.AddDrive(tmpDrive); err != nil {
		Fail(t, tmpDir, err)
	}

	// update the backup directory path for the client, and all
	// files and directories within the client
	newBackupPath := filepath.Join(tmpDir, "new-backup-dir")
	if err := os.Mkdir(newBackupPath, svc.PERMS); err != nil {
		Fail(t, tmpDir, err)
	}
	if err := tmpClient.UpdateBackupPath(newBackupPath); err != nil {
		Fail(t, tmpDir, err)
	}
	// pull all files and directories from temp dbs and verify the
	// backup paths contain newBackupPath
	files, err := tmpClient.Db.GetUsersFiles("me")
	if err != nil {
		Fail(t, tmpDir, err)
	}
	for _, file := range files {
		if !strings.Contains(file.BackupPath, newBackupPath) {
			Fail(t, tmpDir, fmt.Errorf("backup path not found in file object: %s", file.BackupPath))
		}
	}

	dirs, err := tmpClient.Db.GetUsersDirectories("me")
	if err != nil {
		Fail(t, tmpDir, err)
	}
	for _, dir := range dirs {
		if !strings.Contains(dir.BackupPath, newBackupPath) {
			Fail(t, tmpDir, fmt.Errorf("backup path not found in directory: %s", dir.BackupPath))
		}
	}

	if err := Clean(t, tmpDir); err != nil {
		log.Fatal(err)
	}
}

// func TestAddFileToClientAndSendToServer(t *testing.T) {
// 	env.SetEnv(false)
// 	tmpDir, err := envCfgs.Get("CLIENT_TESTING")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// initialize a new testing client
// 	tmpClient := newTestClient(t, tmpDir)
// 	if err := tmpClient.SaveState(); err != nil {
// 		Fail(t, tmpDir, err)
// 	}
// 	tmpClient.SetLocalBackup(false)

// 	// initialize and start a new server in a separate goroutine
// 	stop := make(chan bool)
// 	tmpServer := server.NewServer()
// 	go func() {
// 		tmpServer.Start(stop)
// 	}()

// 	// make and add test file
// 	testFile, err := MakeTmpTxtFile(filepath.Join(tmpClient.Root, "tmp.txt"), RandInt(1000))
// 	if err != nil {
// 		Fail(t, tmpDir, err)
// 	}
// 	if err := tmpClient.AddFile(testFile.ClientPath); err != nil {
// 		Fail(t, tmpDir, err)
// 	}

// 	// send file to server
// 	if err := tmpClient.Push(); err != nil {
// 		Fail(t, tmpDir, err)
// 	}

// 	// stop the server
// 	stop <- true

// 	// pre-emptive cleanup
// 	if err := Clean(t, tmpDir); err != nil {
// 		log.Fatal(err)
// 	}
// }

func TestClientBuildSyncIndex(t *testing.T) {
	env.SetEnv(false)
	tmpDir, err := envCfgs.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}
	// initialize a new client
	tmpClient := newTestClient(t, tmpDir)
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

	drv := svc.NewDrive(auth.NewUUID(), tmpClient.Conf.User, auth.NewUUID(), root.ClientPath, root.ID, root)

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
	tmpDir, err := envCfgs.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}
	// initialize a new client
	tmpClient := newTestClient(t, tmpDir)
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
	tmpClient.Drive.SyncIndex = svc.BuildRootToUpdate(tmpClient.Drive.Root, tmpClient.Drive.SyncIndex)
	assert.NotEqual(t, nil, tmpClient.Drive.SyncIndex.FilesToUpdate)
	assert.NotEqual(t, 0, len(tmpClient.Drive.SyncIndex.FilesToUpdate))
}

func TestClientRefreshDrive(t *testing.T) {
	env.SetEnv(false)
	tmpDir, err := envCfgs.Get("CLIENT_TESTING")
	if err != nil {
		t.Fatal(err)
	}

	// initialize a new client
	tmpClient := newTestClient(t, tmpDir)
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
	tmpDir, err := envCfgs.Get("CLIENT_TESTING")
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
	tmpClient := newTestClient(t, tmpDir)
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
	if err := Clean(t, tmpDir); err != nil {
		log.Fatal(err)
	}
}
