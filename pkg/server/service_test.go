package server

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/env"
	svc "github.com/sfs/pkg/service"

	"github.com/alecthomas/assert/v2"
)

var e = env.NewE()

func getTestingDir() string {
	tmpDir, err := e.Get("SERVICE_TEST_ROOT")
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(tmpDir, "testing")
}

func TestServiceConfig(t *testing.T) {
	env.SetEnv(false)

	c := ServiceConfig()
	assert.NotEqual(t, nil, c)
	assert.NotEqual(t, "", c.SvcRoot)
	assert.NotEqual(t, true, c.IsAdmin)
}

func TestSaveStateFile(t *testing.T) {
	svc := &Service{
		SvcRoot:   GetTestingDir(),
		StateFile: GetTestingDir(),
	}
	// make a tmp folder for the state file
	if err := os.Mkdir(filepath.Join(GetTestingDir(), "state"), 0666); err != nil {
		t.Fatal(err)
	}
	// save state file
	if err := svc.SaveState(); err != nil {
		Fatal(t, err)
	}

	entries, err := os.ReadDir(filepath.Join(GetTestingDir(), "state"))
	if err != nil {
		Fatal(t, err)
	}
	assert.NotEqual(t, 0, len(entries))
	assert.Equal(t, 1, len(entries))
	assert.True(t, strings.Contains(entries[0].Name(), "sfs-state") && strings.Contains(entries[0].Name(), ".json"))

	if err := Clean(GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestLoadServiceFromStateFile(t *testing.T) {
	env.SetEnv(false)
	tmpDir := getTestingDir()
	// create a test service instance with a bunch of sub directories,
	// add to service state, then write out
	svc, err := SetUpService(tmpDir)
	if err != nil {
		Fail(t, tmpDir, err)
	}
	tmpDrv := MakeTmpDriveWithPath(t, tmpDir)
	if err := svc.AddDrive(tmpDrv); err != nil {
		Fail(t, tmpDir, err)
	}

	// load a new test service instance from the newly generated state file
	svc2, err := loadStateFile(svc.StateFile)
	if err != nil {
		Fail(t, tmpDir, err)
	}
	// remove temp state file prior to checking so we don't accidentally
	// leave tmp files behind if an assert fails.
	if err := Clean(filepath.Dir(tmpDir)); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}

	assert.NotEqual(t, nil, svc2)
	assert.Equal(t, svc.SvcRoot, svc2.SvcRoot)
	assert.Equal(t, svc.StateFile, svc2.StateFile)
	assert.Equal(t, svc.UserDir, svc2.UserDir)
	assert.Equal(t, svc.DbDir, svc2.DbDir)
	assert.Equal(t, svc.Users, svc2.Users)
}

func TestLoadServiceFromStateFileAndDbs(t *testing.T) {
	env.SetEnv(false)
	testRoot := getTestingDir()
	testSvc, err := SetUpService(testRoot)
	if err != nil {
		Fatal(t, err)
	}

	// read service entries
	entries, err := os.ReadDir(testSvc.SvcRoot)
	if err != nil {
		t.Fatalf("%v", err)
	}
	assert.NotEqual(t, 0, len(entries))

	// load a new service instance from these resources
	svc, err := SvcLoad(testSvc.SvcRoot)
	if err != nil {
		Fatal(t, err)
	}

	if err := Clean(filepath.Dir(testRoot)); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}

	assert.NotEqual(t, nil, svc)
	assert.Equal(t, testSvc.SvcRoot, svc.SvcRoot)
	assert.Equal(t, testSvc.StateFile, svc.StateFile)
	assert.Equal(t, testSvc.UserDir, svc.UserDir)
	assert.Equal(t, testSvc.DbDir, svc.DbDir)
	assert.Equal(t, testSvc.Users, svc.Users)
}

func TestCreateNewService(t *testing.T) {
	env.SetEnv(false)
	testRoot := getTestingDir()
	testSvc, err := SetUpService(testRoot)
	if err != nil {
		Fatal(t, err)
	}

	entries, err := os.ReadDir(testSvc.SvcRoot)
	if err != nil {
		t.Fatalf("%v", err)
	}
	assert.NotEqual(t, 0, len(entries))

	// check that dbs and state subdirectores aren't empty and
	// the expected files exist. users can be empty since we don't
	// have any yet.
	for _, e := range entries {
		switch e.Name() {
		case "state":
			if isEmpty(filepath.Join(testRoot, "state")) {
				Fatal(t, fmt.Errorf("missing state file directory"))
			}
		case "dbs":
			if isEmpty(filepath.Join(testRoot, "dbs")) {
				Fatal(t, fmt.Errorf("missing database directory"))
			}
		default: // skip anything else for now
			continue
		}
	}
	// clean up
	if err := Clean(filepath.Dir(testRoot)); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

// -------- file tests -------------------------------

// func TestMoveFile(t *testing.T) {
// 	env.SetEnv(false)
// 	testRoot := getTestingDir()
// 	testSvc, err := SetUpService(testRoot)
// 	if err != nil {
// 		Fail(t, filepath.Dir(testRoot), err)
// 	}

// 	// add test drive to test instance
// 	// skipping testSvc.AddDrive() because it creates a new
// 	// drive instance and overwrites testDrv's files map
// 	testDrv := MakeTmpDriveWithPath(t, testRoot)
// 	testSvc.Drives[testDrv.ID] = testDrv
// 	if err := testSvc.Db.AddDrive(testDrv); err != nil {
// 		Fail(t, filepath.Dir(testRoot), err)
// 	}
// 	if err := testSvc.Db.AddDir(testDrv.Root); err != nil {
// 		Fail(t, filepath.Dir(testRoot), err)
// 	}

// 	// pick a file to move
// 	files := testDrv.Root.GetFiles()
// 	file := files[RandInt(len(files)-1)]
// 	file.DriveID = testDrv.ID

// 	// new directory object to represent the temp folder
// 	tmpDir := svc.NewDirectory("new-tmp", "some-rand-id", testDrv.ID, filepath.Join(testRoot, "new-tmp"))
// 	tmpDir.DriveID = testDrv.ID
// 	if err := tmpDir.AddFile(file); err != nil {
// 		Fail(t, filepath.Dir(testRoot), err)
// 	}

// 	if err := testSvc.NewDir(testDrv.ID, testDrv.Root.ID, tmpDir); err != nil {
// 		Fail(t, filepath.Dir(testRoot), err)
// 	}

// 	// attempt to move and then verify the new file path
// 	// with the file object, as well as DB
// 	if err := testSvc.CopyFile(tmpDir.ID, file, false); err != nil {
// 		Fail(t, filepath.Dir(testRoot), err)
// 	}

// 	// verifications
// 	entries, err := os.ReadDir(tmpDir.Path)
// 	if err != nil {
// 		Fail(t, filepath.Dir(testRoot), err)
// 	}
// 	if len(entries) != 1 {
// 		Fail(t, filepath.Dir(testRoot), fmt.Errorf("file was not copied"))
// 	}

// 	// clean up
// 	if err := Clean(filepath.Dir(testRoot)); err != nil {
// 		t.Errorf("[ERROR] unable to remove test directories: %v", err)
// 	}
// }

// ------- user tests --------------------------------

func TestAddAndRemoveUser(t *testing.T) {
	env.SetEnv(false)

	testRoot := getTestingDir()
	testSvc, err := SetUpService(testRoot)
	if err != nil {
		Fail(t, filepath.Dir(testRoot), err)
	}

	// create test user
	testUsr := auth.NewUser("bill buttlicker", "billBB", "bill@bill.com", svcCfg.SvcRoot, false)

	// add user to service instance
	if err := testSvc.AddUser(testUsr); err != nil {
		Fail(t, filepath.Dir(testRoot), err)
	}
	assert.Equal(t, testUsr, testSvc.Users[testUsr.ID])

	// check that its in the db and was entered correctly
	u, err := testSvc.GetUser(testUsr.ID)
	if err != nil {
		Fail(t, filepath.Dir(testRoot), err)
	}
	assert.NotEqual(t, nil, u)
	assert.Equal(t, testUsr.ID, u.ID)

	// attempt to remove user & verify that their drive was removed
	if err := testSvc.RemoveUser(testUsr.ID); err != nil {
		Fail(t, filepath.Dir(testRoot), err)
	}
	entries, err := os.ReadDir(filepath.Join(testSvc.SvcRoot, "users"))
	if err != nil {
		Fail(t, filepath.Dir(testRoot), err)
	}
	// pre-emptively clean incase the asserts fail
	if err := Clean(filepath.Dir(testRoot)); err != nil {
		t.Errorf("[ERROR] unable to remove test user directories: %v", err)
	}

	assert.Equal(t, 0, len(entries))
}

func TestAddAndUpdateAUser(t *testing.T) {
	env.SetEnv(false)
	testRoot := getTestingDir()

	testSvc, err := SetUpService(testRoot)
	if err != nil {
		Fail(t, filepath.Dir(testRoot), err)
	}

	// create a test user
	testUsr := auth.NewUser("bill buttlicker", "billBB", "bill@bill.com", svcCfg.SvcRoot, false)
	if err := testSvc.AddUser(testUsr); err != nil {
		Fail(t, filepath.Dir(testRoot), err)
	}

	// check that its in the db
	u, err := testSvc.GetUser(testUsr.ID)
	if err != nil {
		Fail(t, filepath.Dir(testRoot), err)
	}
	assert.NotEqual(t, nil, u)
	assert.Equal(t, testUsr.ID, u.ID)
	assert.Equal(t, testUsr.Name, u.Name)

	// update name and save
	testUsr.Name = "bill buttlicker II"
	if err := testSvc.UpdateUser(testUsr); err != nil {
		Fail(t, filepath.Dir(testRoot), err)
	}

	// test that the new name is the same as whats in the DB
	u, err = testSvc.GetUser(testUsr.ID)
	if err != nil {
		Fail(t, filepath.Dir(testRoot), err)
	}
	assert.NotEqual(t, nil, u)
	assert.Equal(t, testUsr.ID, u.ID)
	assert.Equal(t, testUsr.Name, u.Name)

	if err := Clean(filepath.Dir(testRoot)); err != nil {
		t.Errorf("[ERROR] unable to clean testing directory: %v", err)
	}
}

// ------ drive tests --------------------------------

func TestAllocateDrive(t *testing.T) {
	env.SetEnv(false)

	testSvc := &Service{
		SvcRoot: GetTestingDir(),
		UserDir: filepath.Join(GetTestingDir(), "users"),
	}

	// create a temp "users" directory
	if err := os.Mkdir(testSvc.UserDir, 0644); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// allocate a new tmp drive
	err := svc.AllocateDrive("test", testSvc.SvcRoot)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// make sure all the basic user files are present
	entries, err := os.ReadDir(GetTestingDir())
	if err != nil {
		Fatal(t, err)
	}
	assert.NotEqual(t, 0, len(entries))

	for _, e := range entries {
		if strings.Contains(e.Name(), "root") {
			continue // don't check root directory
		} else if strings.Contains(e.Name(), "meta") {
			fdir := filepath.Join(GetTestingDir(), "users/test/meta")
			baseFiles, err := os.ReadDir(fdir)
			if err != nil {
				Fatal(t, err)
			}
			for _, bf := range baseFiles {
				assert.True(t, strings.Contains(bf.Name(), ".json"))
			}
		}
	}

	if err := Clean(GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to clean testing directory: %v", err)
	}
}

func TestAddDrive(t *testing.T) {
	env.SetEnv(false)

	testRoot := getTestingDir()
	testSvc, err := SetUpService(testRoot)
	if err != nil {
		Fail(t, filepath.Dir(testRoot), err)
	}

	// test drive
	testDrv := MakeEmptyTmpDrive(t)
	testDrv.Root = nil

	if err := testSvc.AddDrive(testDrv); err != nil {
		Fail(t, filepath.Dir(testRoot), err)
	}

	if err := Clean(filepath.Dir(testRoot)); err != nil {
		t.Errorf("[ERROR] unable to clean testing directory: %v", err)
	}
}

func TestUpdateDrive(t *testing.T) {
	env.SetEnv(false)

	testRoot := getTestingDir()
	testSvc, err := SetUpService(testRoot)
	if err != nil {
		Fail(t, filepath.Dir(testRoot), err)
	}

	// test drive
	testDrv := MakeEmptyTmpDrive(t)

	if err := testSvc.AddDrive(testDrv); err != nil {
		Fail(t, filepath.Dir(testRoot), err)
	}

	// update test drive
	testDrv.OwnerName = "william j buttlicker"

	if err := testSvc.UpdateDrive(testDrv); err != nil {
		Fail(t, filepath.Dir(testRoot), err)
	}

	// verify the name of the owner of the drive
	drv, err := testSvc.Db.GetDrive(testDrv.ID)
	if err != nil {
		Fail(t, filepath.Dir(testRoot), err)
	}
	if drv.OwnerName != testDrv.OwnerName {
		Fail(t, filepath.Dir(testRoot), fmt.Errorf("owner name mismatch. orig: %s new: %s", testDrv.OwnerName, drv.OwnerName))
	}

	if err := Clean(filepath.Dir(testRoot)); err != nil {
		t.Errorf("[ERROR] unable to clean testing directory: %v", err)
	}
}

func TestLoadDrive(t *testing.T) {
	env.SetEnv(false)

	testRoot := getTestingDir()
	testSvc, err := SetUpService(testRoot)
	if err != nil {
		Fail(t, filepath.Dir(testRoot), err)
	}

	testDrv := MakeEmptyTmpDrive(t)
	if err := testSvc.AddDrive(testDrv); err != nil {
		Fail(t, filepath.Dir(testRoot), err)
	}

	foundDrv, err := testSvc.LoadDrive(testDrv.ID)
	if err != nil {
		Fail(t, filepath.Dir(testRoot), err)
	}

	if err := Clean(filepath.Dir(testRoot)); err != nil {
		t.Errorf("[ERROR] unable to clean testing directory: %v", err)
	}

	assert.NotEqual(t, nil, foundDrv)
	assert.Equal(t, testDrv.ID, foundDrv.ID)
}

// func TestRemoveDrive(t *testing.T) {}

// func TestServiceReset(t *testing.T) {
// 	env.SetEnv(false)

// 	// create a test service in admin mode (since reset requires admin)
// 	testSvc, err := Init(false, true)
// 	if err != nil {
// 		Fail(t, GetTestingDir(), err)
// 	}
// 	assert.True(t, testSvc.AdminMode)

// 	// add a bunch of test users.
// 	// testSvc will allocate drives for each test user.
// 	for i := 0; i < 10; i++ {
// 		usersDir := filepath.Join(testSvc.UserDir, fmt.Sprintf("bill-%d", i+1))
// 		user := auth.NewUser(fmt.Sprintf("bill-%d", i+1), "billderper", "derper@derp.com", usersDir, false)
// 		if err := testSvc.AddUser(user); err != nil {
// 			Fail(t, testSvc.SvcRoot, err)
// 		}
// 	}

// 	// run reset, then verify the sfs/users folder is empty,
// 	// and that each of the tables no longer have the test users
// 	// that were just generated
// 	if err := testSvc.Reset(testSvc.SvcRoot); err != nil {
// 		Fail(t, testSvc.SvcRoot, err)
// 	}

// 	if err := Clean(testSvc.UserDir); err != nil {
// 		t.Fatal(err)
// 	}
// }
