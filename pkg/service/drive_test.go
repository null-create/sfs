package service

import (
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
)

// make a dummy file system. does NOT create assoiatedfiles!
// contains only mocked internal data structures
func MakeDummySystem(t *testing.T) (*Drive, *Directory) {
	testingDir := GetTestingDir()
	driveRoot := filepath.Join(testingDir, "testDrive")

	// make a tmp directory and add with some test files
	if _, err := MakeTmpDir(t, driveRoot); err != nil {
		t.Errorf("[ERROR] failed to create temporary directory: %v", err)
	}

	testFiles1 := MakeTestDirFiles(t, 5, driveRoot)
	testRoot := NewRootDirectory("testRoot", "me", driveRoot)
	testRoot.AddFiles(testFiles1)

	testDrive := NewDrive(NewUUID(), "testDrive", "me", driveRoot, NewUUID(), testRoot)

	// create a subdirectory with dummy files
	sdPath := filepath.Join(driveRoot, "testSubDir")
	if _, err := MakeTmpDir(t, sdPath); err != nil {
		t.Errorf("[ERROR] unable to create temporary subdirectory: %v", err)
	}

	testFiles2 := MakeTestDirFiles(t, 5, sdPath)
	testDirectory := NewDirectory("test-dir", "me", sdPath)
	testDirectory.AddFiles(testFiles2)

	testRoot.AddSubDir(testDirectory)

	return testDrive, testRoot
}

// ----------------------------------------------------------------

func TestDriveSecurityFeatures(t *testing.T) {
	testDir := GetTestingDir()
	tmpDir := filepath.Join(testDir, "testDir")

	testRoot := NewRootDirectory("testRoot", "me", filepath.Join(tmpDir, "testRoot"))
	testDrive := NewDrive(NewUUID(), "test-drive", "me", tmpDir, NewUUID(), testRoot)

	testDrive.Lock("default")
	assert.True(t, testDrive.Protected)

	testDrive.SetNewPassword("wrongPassword", "newPassword", false)
	assert.NotEqual(t, "wrongPassword", testDrive.Key)

	testDrive.Unlock("default")
	testDrive.SetNewPassword("default", "newPassword", false)
	assert.Equal(t, "newPassword", testDrive.Key)

	// test admin override
	testDrive.Lock("newPassword")
	assert.True(t, testDrive.Protected)

	testDrive.SetNewPassword("anotherPassword", "newPassword", true)
	assert.Equal(t, "newPassword", testDrive.Key)
}
