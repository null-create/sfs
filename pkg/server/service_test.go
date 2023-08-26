package server

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
)

// --------- fixtures --------------------------------

func fakeServiceRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("[ERROR] failed to get current working directory: %v", err)
	}
	return cwd
}

// -----------------------------------------------------

func TestSaveStateFile(t *testing.T) {
	svc := &Service{}
	svc.SvcRoot = fakeServiceRoot()
	svc.StateFile = GetStateDir()
	if err := svc.SaveState(); err != nil {
		Fatal(t, err)
	}

	entries, err := os.ReadDir(GetStateDir())
	if err != nil {
		Fatal(t, err)
	}
	assert.NotEqual(t, 0, len(entries))
	assert.Equal(t, 1, len(entries))
	assert.True(t, strings.Contains(entries[0].Name(), "sfs-state") && strings.Contains(entries[0].Name(), ".json"))

	if err := Clean(t, GetStateDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestLoadFromStateFile(t *testing.T) {
	svc := &Service{}
	svc.SvcRoot = fakeServiceRoot()
	svc.StateFile = GetStateDir()

	if err := svc.SaveState(); err != nil {
		Fatal(t, err)
	}

	svc2, err := loadStateFile(svc.StateFile)
	if err != nil {
		Fatal(t, err)
	}

	// remove temp state file prior to checking so we don't accidentally
	// leave tmp files behind
	if err := Clean(t, filepath.Join(fakeServiceRoot(), "state")); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}

	assert.NotEqual(t, nil, svc2)
}

func TestGenBaseUserFiles(t *testing.T) {
	svc := &Service{}
	svc.SvcRoot = GetTestingDir()
	svc.UserDir = filepath.Join(svc.SvcRoot, "users")

	svc.GenBaseUserFiles(svc.UserDir)

	entries, err := os.ReadDir(svc.UserDir)
	if err != nil {
		Fatal(t, err)
	}
	assert.NotEqual(t, 0, len(entries))
	for _, e := range entries {
		assert.True(t, strings.Contains(e.Name(), ".json"))
	}

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

// func TestAllocateDrive(t *testing.T) {}

// func TestCreateNewService(t *testing.T) {
// 	// use service.SvcInit(path string)

// }

// func TestSvcUserGets(t *testing.T) {}

// func TestSvcUserRemoves(t *testing.T) {}

// func TestSvcClearAll(t *testing.T) {}
