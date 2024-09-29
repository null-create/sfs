package service

import (
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/env"
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
	testRoot := NewRootDirectory("testRoot", "some-rand-id", "some-rand-id", driveRoot)
	testRoot.AddFiles(testFiles1)

	testDrive := NewDrive(auth.NewUUID(), "testDrive", "me", driveRoot, auth.NewUUID(), testRoot)

	// create a subdirectory with dummy files
	sdPath := filepath.Join(driveRoot, "testSubDir")
	if _, err := MakeTmpDir(t, sdPath); err != nil {
		t.Errorf("[ERROR] unable to create temporary subdirectory: %v", err)
	}

	testFiles2 := MakeTestDirFiles(t, 5, sdPath)
	testDirectory := NewDirectory("test-dir", "me", "some-rand-id", sdPath)
	testDirectory.AddFiles(testFiles2)

	testRoot.AddSubDir(testDirectory)

	return testDrive, testRoot
}

// ----------------------------------------------------------------

func TestDriveSecurityFeatures(t *testing.T) {
	env.SetEnv(false)

	testDir := GetTestingDir()
	tmpDir := filepath.Join(testDir, "testDir")

	testRoot := NewRootDirectory("testRoot", "some-rand-id", "some-rand-id", filepath.Join(tmpDir, "testRoot"))
	testDrive := NewDrive(auth.NewUUID(), "test-drive", "me", tmpDir, auth.NewUUID(), testRoot)

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

func TestPopulateDrive(t *testing.T) {
	env.SetEnv(false)

	// make test files without creating file/dir objects
	MakeTmpDirs(t)

	// create a test root & populate with test files
	tmpDrive := MakeEmptyTmpDrive(t)
	tmpDrive.Root = tmpDrive.Root.Walk()

	// clean up test files
	if err := Clean(t, GetTestingDir()); err != nil {
		t.Fatal(err)
	}
}
