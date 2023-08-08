package files

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
)

// make a dummy file system. does NOT create assoiatedfiles!
// contains only mocked internal data structures
func MakeDummySystem(t *testing.T) (*Drive, *Directory) {
	testingDir := GetTestingDir()
	driveRoot := filepath.Join(testingDir, "testDir")

	// create test drive and a dummy user root directory
	testRoot := NewRootDirectory("testRoot", "me", driveRoot)
	testFiles1 := MakeTestDirFiles(t, 5, testRoot.Path)
	testRoot.AddFiles(testFiles1)

	testDrive := NewDrive(NewUUID(), "testDrive", "me", driveRoot, testRoot)

	// create a subdirectory with dummy files
	testDirectory := NewDirectory("test-dir", "me", filepath.Join(driveRoot, "testSubDir"))
	testFiles2 := MakeTestDirFiles(t, 5, testDirectory.Path)
	testDirectory.AddFiles(testFiles2)

	testRoot.AddSubDir(testDirectory)

	return testDrive, testRoot
}

// removes only the actual test files, directories and subdirectories directly
// under ../nimbus/pkg/files/test_files
func RemoveDummySystem(t *testing.T) (err error) {
	testingDir := GetTestingDir()
	driveRoot := filepath.Join(testingDir, "testDir")

	if err := os.Remove(driveRoot); err != nil {
		return err
	}

	return nil
}

// ----------------------------------------------------------------

func TestDriveSecurityFeatures(t *testing.T) {
	testDir := GetTestingDir()
	tmpDir := filepath.Join(testDir, "testDir")

	testRoot := NewRootDirectory("testRoot", "me", filepath.Join(tmpDir, "testRoot"))
	testDrive := NewDrive(NewUUID(), "test-drive", "me", tmpDir, testRoot)

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

func TestGetDriveSize(t *testing.T) {
	testDrive, _ := MakeDummySystem(t)

	if testFiles, err := MakeTestFiles(t, 10); err == nil {
		testDrive.Root.AddFiles(testFiles)
	} else {
		t.Errorf("[ERROR] unable to make test files: %v", err)
	}
	assert.NotEqual(t, 0, len(testDrive.Root.Files))

	// get the size of the drive
	dirSize := testDrive.DriveSize()
	assert.NotEqual(t, 0, dirSize)

	// TODO: figure out actual expected size to compare to

	if err := RemoveTestFiles(t, 10); err != nil {
		t.Errorf("[ERROR] unable to remove test files: %v", err)
	}
}
