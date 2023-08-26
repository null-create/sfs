package server

import (
	"os"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestSaveStateFile(t *testing.T) {
	// use Service.SaveState()
	svc := &Service{}
	svc.SfDir = GetTestingDir()
	if err := svc.SaveState(); err != nil {
		t.Fatalf("%v", err)
	}

	entries, err := os.ReadDir(GetTestingDir())
	if err != nil {
		t.Fatalf("%v", err)
	}
	assert.NotEqual(t, 0, len(entries))
	assert.True(t, strings.Contains(entries[0].Name(), "sfs-state") && strings.Contains(entries[0].Name(), ".json"))

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestLoadStateFile(t *testing.T) {}

func TestCreateNewService(t *testing.T) {
	// use service.SvcInit(path string)

}

func TestLoadServiceFromStateFile(t *testing.T) {
	// use service.SvcLoad

}

func TestGenBaseUserFiles(t *testing.T) {}

func TestAllocateDrive(t *testing.T) {}

func TestSvcUserGets(t *testing.T) {}

func TestSvcUserRemoves(t *testing.T) {}

func TestSvcClearAll(t *testing.T) {}
