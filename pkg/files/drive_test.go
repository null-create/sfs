package files

import (
	"os"
	"testing"

	"github.com/alecthomas/assert/v2"
)

// make a dummy file system. does NOT create assoiatedfiles!
// contains only mocked internal data structures
func MakeDummySystem(t *testing.T) (*Drive, *Directory, *Directory) {

	testDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("[ERROR] failed to get current working directory: %v", err)
	}

	testDrive := NewDrive("testDrive", "test-drive", "me", "/")
	testRoot := NewRootDirectory("testRoot", "me", "path/to/test.txt")
	testDirectory := NewDirectory("test-dir", "me", testDir)
	testDirectory.Parent = testRoot

	return testDrive, testRoot, testDirectory
}

func TestDriveSecurityFeatures(t *testing.T) {
	testDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("[ERROR] failed to get current working directory: %v", err)
	}

	testDrive := NewDrive("1", "test-drive", "me", testDir)

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

// func TestGetDriveSize(t *testing.T) {
// 	total := RandInt(MAX)
// 	MakeTestFiles(t, total)

// 	testDrive, _, _ := MakeDummySystem(t)
// 	testDrive.Root.AddFiles(MakeDummyFiles(t, total))

// 	// iterate over directory and get the total size,
// 	// assert result is not 0. maybe find a way to get actual size
// 	// and make sure result is equal as well.

// 	RemoveTestFiles(t, total)
// }
