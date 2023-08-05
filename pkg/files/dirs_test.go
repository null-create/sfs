package files

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
)

const (
	TestDirName = "testDir"
)

// makes test subdirectories
//
// (and eventually actual subdirs within nimbus/pkg/files/test_files)
func MakeTestDirs(t *testing.T, total int) []*Directory {
	testingDir := GetTestingDir()

	testDirs := make([]*Directory, 0)
	for i := 0; i < total; i++ {
		tdName := fmt.Sprintf("%s%d", TestDirName, i)
		tmpDir := filepath.Join(testingDir, tdName)

		// if err := os.Mkdir(tmpDir, 0666); err != nil {
		// 	t.Errorf("[ERROR] failed to create temporary directory: %v", err)
		// }

		testDirs = append(testDirs, NewDirectory(tdName, "me", tmpDir))
	}

	return testDirs
}

// currently a no-op until the access issue is resolved
func RemoveTestDirs(t *testing.T, total int) error {
	testingDir := GetTestingDir()

	entries, err := os.ReadDir(testingDir)
	if err != nil {
		t.Errorf("[ERROR] failed to read testing directory: %v", err)
	}

	for _, entry := range entries {
		path := filepath.Join(testingDir, entry.Name())
		if entry.IsDir() {
			// remove the directory
			// NOTE: currently getting "access denied" errors when
			// removing temp directories
			if err = os.Remove(path); err != nil {
				t.Errorf("[ERROR] failed to remove temp directory: %v", err)
			}
		}
	}
	return nil
}

func TestSecurityFeatures(t *testing.T) {
	testDirs := MakeTestDirs(t, 2)
	td := testDirs[0]
	td2 := testDirs[1]

	testFiles, err := MakeTestFiles(t, 1)
	if err != nil {
		t.Fatalf("[ERROR] failed to create test files: %v", err)
	}
	tf := testFiles[0]

	// lock test directory and attempt to add a file and subdirectory
	td.Lock("default")

	td.AddSubDir(td2)
	assert.Equal(t, 0, len(td.Dirs))
	td.AddFile(tf)
	assert.Equal(t, 0, len(td.Files))

	// unlock, add test dir and test file, then re-lock and try to remove
	td.Unlock("default")
	td.AddFile(tf)
	td.AddSubDir(td2)
	td.Lock("default")

	td.RemoveFile(tf.ID)
	td.RemoveSubDir(td2.ID)
	assert.NotEqual(t, 0, len(td.Dirs))
	assert.NotEqual(t, 0, len(td.Files))

	// attempt to change password, then re-lock and try to remove
	td.SetPassword("wrongPassword", "newPassword")
	assert.NotEqual(t, "wrongPassword", td.Key)
	td.SetPassword("default", "newPassword")
	assert.Equal(t, "newPassword", td.Key)

	td.Unlock("newPassword")
	td.AddFile(tf)
	td.AddSubDir(td2)
	td.Lock("newPassword")

	td.RemoveFile(tf.ID)
	td.RemoveSubDir(td2.ID)
	assert.NotEqual(t, 0, len(td.Dirs))
	assert.NotEqual(t, 0, len(td.Files))

	if err := RemoveTestDirs(t, 2); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestAddFiles(t *testing.T) {
	total := RandInt(10)
	testDirs := MakeTestDirs(t, 1)
	assert.Equal(t, 0, len(testDirs[0].Files))

	// make some test files
	testFiles, err := MakeTestFiles(t, total)
	if err != nil {
		t.Errorf("[ERROR] unable to make test files: %v", err)
	}

	td := testDirs[0]
	for _, testFile := range testFiles {
		td.AddFile(testFile)
	}
	assert.NotEqual(t, 0, len(td.Files))
	assert.Equal(t, total, len(td.Files))

	if err := RemoveTestDirs(t, 1); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
	if err := RemoveTestFiles(t, total); err != nil {
		t.Errorf("[ERROR] unable to remove test files: %v", err)
	}
}

func TestRemoveFiles(t *testing.T) {
	total := RandInt(10)
	testDirs := MakeTestDirs(t, 1)
	assert.Equal(t, 0, len(testDirs[0].Files))

	// make some test files
	testFiles, err := MakeTestFiles(t, total)
	if err != nil {
		t.Errorf("[ERROR] unable to make test files: %v", err)
	}

	td := testDirs[0]
	for _, testFile := range testFiles {
		td.AddFile(testFile)
	}
	assert.NotEqual(t, 0, len(td.Files))

	// remove test files
	for _, testFile := range testFiles {
		td.RemoveFile(testFile.ID)
	}
	assert.Equal(t, 0, len(td.Files))

	if err := RemoveTestDirs(t, 1); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
	if err := RemoveTestFiles(t, total); err != nil {
		t.Errorf("[ERROR] unable to remove test files: %v", err)
	}

}

func TestAddSubDirs(t *testing.T) {
	total := RandInt(100)                // root test subdir
	testDirs := MakeTestDirs(t, total+1) // subdirs to add + 1 for a test root

	td := testDirs[0]
	testDirs = append(testDirs[:0], testDirs[1:]...) // remove test dir from original slice

	if err := td.AddSubDirs(testDirs); err != nil {
		t.Errorf("[ERROR] unable to add subdirs to test directory: %v", err)
	}
	assert.Equal(t, total, len(td.Dirs))

	if err := RemoveTestDirs(t, total); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestRemoveSubDirs(t *testing.T) {
	total := RandInt(100)                // root test subdir
	testDirs := MakeTestDirs(t, total+1) // subdirs to add + 1 for a test root

	td := testDirs[0]
	testDirs = append(testDirs[:0], testDirs[1:]...) // remove test dir from original slice

	if err := td.AddSubDirs(testDirs); err != nil {
		t.Errorf("[ERROR] unable to add subdirs to test directory: %v", err)
	}

	for _, testDir := range testDirs {
		if err := td.RemoveSubDir(testDir.ID); err != nil {
			t.Errorf("[ERROR] unable to remove test subdir: %v", err)
		}
	}
	assert.Equal(t, 0, len(td.Dirs))
}

func TestGetDirSize(t *testing.T) {
	testDir := NewDirectory("testDir", "me", GetTestingDir())
	tf, err := MakeTestFiles(t, 1)
	if err != nil {
		t.Errorf("[ERROR] unable to create test files: %v", err)
	}
	testDir.AddFile(tf[0])

	tdSize, err := testDir.DirSize()
	if err != nil {
		t.Errorf("[ERROR] unable to to get directory size %v", err)
	}
	assert.NotEqual(t, 0, tdSize)

	if err := RemoveTestFiles(t, 1); err != nil {
		t.Errorf("[ERROR] unable to remove test file %v", err)
	}
}

func TestWalk(t *testing.T) {}

func TestWalkS(t *testing.T) {}

func TestWalkF(t *testing.T) {}
