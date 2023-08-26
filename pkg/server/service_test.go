package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sfs/pkg/auth"

	"github.com/alecthomas/assert/v2"
)

// --------- fixtures --------------------------------

func MakeABunchOfDummyUsers(total int) []*auth.User {
	return nil
}

// -----------------------------------------------------

func TestSaveStateFile(t *testing.T) {
	svc := &Service{}
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
	svc.InitTime = time.Now().UTC()
	svc.SvcRoot = GetTestingDir()
	svc.StateFile = GetStateDir()

	if err := svc.SaveState(); err != nil {
		t.Fatalf("%v", err)
	}

	svc2, err := loadStateFile(svc.StateFile)
	if err != nil {
		Fatal(t, err)
	}

	// remove temp state file prior to checking so we don't accidentally
	// leave tmp files behind
	if err := Clean(t, filepath.Join(GetTestingDir(), "state")); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}

	assert.NotEqual(t, nil, svc2)
}

// func TestCreateNewService(t *testing.T) {
// 	// use service.SvcInit(path string)

// }

// func TestLoadServiceFromStateFile(t *testing.T) {
// 	// use service.SvcLoad

// }

// func TestGenBaseUserFiles(t *testing.T) {}

// func TestAllocateDrive(t *testing.T) {}

// func TestSvcUserGets(t *testing.T) {}

// func TestSvcUserRemoves(t *testing.T) {}

// func TestSvcClearAll(t *testing.T) {}
