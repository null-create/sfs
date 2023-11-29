package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sfs/pkg/auth"

	"github.com/alecthomas/assert/v2"
)

func TestServiceConfig(t *testing.T) {
	BuildEnv(true)
	c := ServiceConfig()
	assert.NotEqual(t, nil, c)
	assert.True(t, strings.Contains(c.SvcRoot, "C:"))
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
	// clean up after any previous failed runs
	if err := Clean(GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
	// create a test service instance
	svc := &Service{
		SvcRoot:   GetTestingDir(),
		StateFile: GetTestingDir(),
	}
	// make a tmp folder for the state file
	if err := os.Mkdir(filepath.Join(GetTestingDir(), "state"), 0666); err != nil {
		Fatal(t, err)
	}
	if err := svc.SaveState(); err != nil {
		Fatal(t, err)
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
	testSvc, err := SvcInit(testRoot, true)
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
	svc, err := SvcLoad(testSvc.SvcRoot, true)
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
	testSvc, err := SvcInit(testRoot, true)
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

func TestAllocateDrive(t *testing.T) {
	svc := &Service{
		SvcRoot: GetTestingDir(),
		UserDir: filepath.Join(GetTestingDir(), "users"),
	}

	// create a temp "users" directory
	if err := os.Mkdir(svc.UserDir, 0644); err != nil {
		t.Fatalf("%v", err)
	}

	// allocate a new tmp drive
	d, err := AllocateDrive("test", "me", svc.SvcRoot)
	if err != nil {
		Fatal(t, err)
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
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestAddAndRemoveUser(t *testing.T) {
	BuildEnv(true)

	// create test service instance
	conf := ServiceConfig()
	testFolder := filepath.Join(conf.SvcRoot, "users")
	testSvc, err := SvcLoad(conf.SvcRoot, true)
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
	testDrv, err := testSvc.FindDrive(testUsr.DriveID)
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
	BuildEnv(true)
	c := ServiceConfig()
	// create a test instance
	testSvc, err := SvcLoad(c.SvcRoot, true)
	if err != nil {
		t.Fatal(err)
	}
	// create a test user
	testUsr := auth.NewUser("bill buttlicker", "billBB", "bill@bill.com", c.SvcRoot, false)
	if err := testSvc.AddUser(testUsr); err != nil {
		t.Fatal(err)
	}
	origName := testUsr.Name

	// check that its in the db
	u, err := testSvc.FindUser(testUsr.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEqual(t, nil, u)
	assert.Equal(t, testUsr.ID, u.ID)
	assert.Equal(t, testUsr.Name, u.Name)

	// update name and save
	testUsr.Name = "bill buttlicker II"
	if err := testSvc.UpdateUser(testUsr); err != nil {
		t.Fatal(err)
	}

	// test that the new name is the same as whats in the DB
	u, err = testSvc.FindUser(testUsr.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEqual(t, nil, u)
	assert.Equal(t, testUsr.ID, u.ID)
	assert.Equal(t, testUsr.Name, u.Name)

	if err := Clean(GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
	// remove test user drive
	tmpDir := filepath.Join(c.SvcRoot, "users", origName)
	if err := Clean(tmpDir); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

// func TestAddDrive(t *testing.T) {}

// func TestAddAndUpdateDrive(t *testing.T) {}

// func TestAddAndRemoveDrive(t *testing.T) {}
