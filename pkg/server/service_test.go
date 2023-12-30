package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/env"
	svc "github.com/sfs/pkg/service"

	"github.com/alecthomas/assert/v2"
)

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

	// create a test service instance with a bunch of sub directories,
	// add to service state, then write out
	svc, err := Init(false, false)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}
	tmpDrv := MakeTmpDrive(t)
	if err := svc.AddDrive(tmpDrv); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// load a new test service instance from the newly generated state file
	svc2, err := loadStateFile(svc.StateFile)
	if err != nil {
		Fatal(t, err)
	}
	// remove temp state file prior to checking so we don't accidentally
	// leave tmp files behind if an assert fails.
	if err := Clean(GetTestingDir()); err != nil {
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
	testRoot := filepath.Join(GetTestingDir(), "tmp")
	testSvc, err := SvcInit(testRoot)
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
	assert.NotEqual(t, nil, svc)
	assert.Equal(t, testSvc.SvcRoot, svc.SvcRoot)
	assert.Equal(t, testSvc.StateFile, svc.StateFile)
	assert.Equal(t, testSvc.UserDir, svc.UserDir)
	assert.Equal(t, testSvc.DbDir, svc.DbDir)
	assert.Equal(t, testSvc.Users, svc.Users)

	if err := Clean(GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestCreateNewService(t *testing.T) {
	// clean up after any previous failed runs
	if err := Clean(GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}

	testRoot := filepath.Join(GetTestingDir(), "tmp")
	testSvc, err := SvcInit(testRoot)
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
	if err := Clean(GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestAddAndRemoveUser(t *testing.T) {
	env.SetEnv(false)

	// create test service instance
	conf := ServiceConfig()
	testFolder := filepath.Join(conf.SvcRoot, "users")
	testSvc, err := SvcLoad(conf.SvcRoot)
	if err != nil {
		Fail(t, testFolder, err)
	}

	// create test user
	testUsr := auth.NewUser("bill buttlicker", "billBB", "bill@bill.com", conf.SvcRoot, false)

	// add user to service instance
	if err := testSvc.AddUser(testUsr); err != nil {
		Fail(t, testFolder, err)
	}
	assert.Equal(t, testUsr, testSvc.Users[testUsr.ID])

	// check that its in the db and was entered correctly
	u, err := testSvc.FindUser(testUsr.ID)
	if err != nil {
		Fail(t, testFolder, err)
	}
	assert.NotEqual(t, nil, u)
	assert.Equal(t, testUsr.ID, u.ID)

	// get a tmp drive to check that things have been removed correctly
	testDrv, err := testSvc.LoadDrive(testUsr.DriveID)
	if err != nil {
		Fail(t, testFolder, err)
	}
	assert.NotEqual(t, nil, testDrv)

	// attempt to remove user & verify that their drive was removed
	if err := testSvc.RemoveUser(testUsr.ID); err != nil {
		Fail(t, testFolder, err)
	}
	entries, err := os.ReadDir(filepath.Join(conf.SvcRoot, "users"))
	if err != nil {
		Fail(t, testFolder, err)
	}
	assert.Equal(t, 0, len(entries))

	if err := Clean(filepath.Join(conf.SvcRoot, "users")); err != nil {
		t.Errorf("[ERROR] unable to remove test user directories: %v", err)
	}
}

func TestAddAndUpdateAUser(t *testing.T) {
	env.SetEnv(false)

	c := ServiceConfig()
	// create a test instance
	testSvc, err := SvcLoad(c.SvcRoot)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}
	// create a test user
	testUsr := auth.NewUser("bill buttlicker", "billBB", "bill@bill.com", c.SvcRoot, false)
	if err := testSvc.AddUser(testUsr); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// check that its in the db
	u, err := testSvc.FindUser(testUsr.ID)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}
	assert.NotEqual(t, nil, u)
	assert.Equal(t, testUsr.ID, u.ID)
	assert.Equal(t, testUsr.Name, u.Name)

	// update name and save
	testUsr.Name = "bill buttlicker II"
	if err := testSvc.UpdateUser(testUsr); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// test that the new name is the same as whats in the DB
	u, err = testSvc.FindUser(testUsr.ID)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}
	assert.NotEqual(t, nil, u)
	assert.Equal(t, testUsr.ID, u.ID)
	assert.Equal(t, testUsr.Name, u.Name)

	if err := Clean(GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to clean testing directory: %v", err)
	}
	// remove test user drive
	tmpDir := filepath.Join(c.SvcRoot, "users")
	if err := Clean(tmpDir); err != nil {
		t.Errorf("[ERROR] unable to clean up test user files: %v", err)
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
	d, err := svc.AllocateDrive("test", "me", testSvc.SvcRoot)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}
	assert.NotEqual(t, nil, d)

	// make sure all the basic user files are present
	entries, err := os.ReadDir(d.DriveRoot)
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

	// test service
	testSvc, err := Init(false, false)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// test drive
	testDrv := MakeEmptyTmpDrive(t)

	if err := testSvc.AddDrive(testDrv); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	if err := Clean(GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to clean testing directory: %v", err)
	}
}

func TestUpdateDrive(t *testing.T) {
	env.SetEnv(false)

	// test service
	testSvc, err := Init(true, false)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// test drive
	testDrv := MakeEmptyTmpDrive(t)

	if err := testSvc.AddDrive(testDrv); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// update test drive
	testDrv.OwnerName = "william j buttlicker"

	if err := testSvc.UpdateDrive(testDrv); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	if err := Clean(GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to clean testing directory: %v", err)
	}
}

// func TestRefreshDrive(t *testing.T) {}

// func TestRemoveDrive(t *testing.T) {}

func TestDiscover(t *testing.T) {
	env.SetEnv(false)

	testSvc, err := Init(false, false)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}

	MakeTmpDirs(t)
	tmpDrive := MakeEmptyTmpDrive(t)

	tmpDrive.Root, err = testSvc.Discover(tmpDrive.Root)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// retrieve files and confirm that they exist
	tmpRootDirs := tmpDrive.Root.WalkDs()
	if len(tmpRootDirs) == 0 {
		Fail(t, GetTestingDir(), fmt.Errorf("no test directories found"))
	}
	tmpRootFiles := tmpDrive.Root.WalkFs()
	if len(tmpRootFiles) == 0 {
		Fail(t, GetTestingDir(), fmt.Errorf("no test directories found"))
	}

	if err := Clean(GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestPopulate(t *testing.T) {
	env.SetEnv(false)

	testSvc, err := Init(false, false)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// create a test drive with files and directories
	// added to the database via Discover()
	tmpDrive := MakeTmpDrive(t)
	tmpDrive.Root, err = testSvc.Discover(tmpDrive.Root)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// create a new root directory and populate
	testRoot := svc.NewRootDirectory("test", "some-rand-id", tmpDrive.Root.ID, filepath.Join(GetTestingDir(), "tmp"))
	testRoot = testSvc.Populate(testRoot)

	testRootFiles := testRoot.WalkFs()
	if len(testRootFiles) == 0 {
		Fail(t, GetTestingDir(), fmt.Errorf("failed to populate directories"))
	}
	testRootDirs := testRoot.WalkDs()
	if len(testRootDirs) == 0 {
		Fail(t, GetTestingDir(), fmt.Errorf("failed to populate files"))
	}

	if err := Clean(GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

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
