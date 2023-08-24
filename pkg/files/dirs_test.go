package files

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
)

const TestDirName string = "testDir"

//---------- test fixtures --------------------------------

// creates an empty directory under ../nimbus/pkg/files/test_files
func MakeTmpDir(t *testing.T, path string) error {
	if err := os.Mkdir(path, 0666); err != nil {
		return fmt.Errorf("[ERROR] unable to create temporary directory: %v", err)
	}
	return nil
}

// make test files within a testing/tmp directory
func MakeTestDirFiles(t *testing.T, total int, tdPath string) []*File {
	testFiles := make([]*File, 0)

	for i := 0; i < total; i++ {
		name := fmt.Sprintf("test-file-%d.txt", i)
		tfPath := filepath.Join(tdPath, name)

		// Create creates or truncates the named file.
		file, err := os.Create(tfPath)
		if err != nil {
			t.Fatalf("[ERROR] failed to create test file: %v", err)
		}
		file.Write([]byte(testData))
		file.Close()

		testFiles = append(testFiles, NewFile(name, "me", tfPath))
	}

	return testFiles
}

// makes testing/tmp directory objects.
//
// *does not* create actual test directories.
// this is typically done via directory.AddSubDir()
func MakeTestDirs(t *testing.T, total int) []*Directory {
	testingDir := GetTestingDir()

	testDirs := make([]*Directory, 0)
	for i := 0; i < total; i++ {
		tdName := fmt.Sprintf("%s%d", TestDirName, i)
		tmpDirPath := filepath.Join(testingDir, tdName)

		// NOTE: none of these test directories have a non-nil parent pointer
		testDirs = append(testDirs, NewDirectory(tdName, "me", tmpDirPath))
	}
	return testDirs
}

func RemoveTmpDir(t *testing.T, path string) error {
	if err := os.Remove(path); err != nil {
		t.Errorf("[ERROR] unable to remove temp directory: %v", err)
	}
	return nil
}

// ----------------------------------------------------------------

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

	Clean(t, GetTestingDir())
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

	Clean(t, GetTestingDir())
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

	Clean(t, GetTestingDir())
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

	if err := Clean(t, GetTestingDir()); err != nil {
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
	assert.Equal(t, total, len(td.Dirs))

	if err := td.RemoveSubDir(td.ID); err != nil {
		t.Errorf("[ERROR] unable to remove test subdir: %v", err)
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

func TestWalk(t *testing.T) {
	testingDir := GetTestingDir()
	testDir1 := NewDirectory("testDir1", "me", filepath.Join(testingDir, "testDir1"))
	testDir2 := NewDirectory("testDir2", "me", filepath.Join(testingDir, "testDir2"))
	testDir3 := NewDirectory("testDir3", "me", filepath.Join(testingDir, "testDir3"))
	testDir4 := NewDirectory("testDir4", "me", filepath.Join(testingDir, "testDir4"))
	testDir5 := NewDirectory("testDir5", "me", filepath.Join(testingDir, "testDir5"))

	idToFind := testDir3.ID

	testDirs := []*Directory{testDir1, testDir2, testDir3, testDir4, testDir5}

	// add a bunch of dummy directories to each testDir
	for _, td := range testDirs {
		if err := td.AddSubDirs(MakeTestDirs(t, RandInt(50))); err != nil {
			t.Errorf("[ERROR] unable to add test directories %v", err)
		}
	}

	// make layered file system
	testDir5.AddSubDir(testDir4)
	testDir4.AddSubDir(testDir3)
	testDir3.AddSubDir(testDir2)
	testDir2.AddSubDir(testDir1)

	// run Walk and check result
	dir := testDir5.Walk(idToFind)
	assert.NotEqual(t, nil, dir)
	assert.Equal(t, idToFind, dir.ID)
}

func TestWalkS(t *testing.T) {
	testingDir := GetTestingDir()
	testDir1 := NewDirectory("testDir1", "me", filepath.Join(testingDir, "testDir1"))
	testDir2 := NewDirectory("testDir2", "me", filepath.Join(testingDir, "testDir2"))
	testDir3 := NewDirectory("testDir3", "me", filepath.Join(testingDir, "testDir3"))
	testDir4 := NewDirectory("testDir4", "me", filepath.Join(testingDir, "testDir4"))
	testDir5 := NewDirectory("testDir5", "me", filepath.Join(testingDir, "testDir5"))

	testDirs := []*Directory{testDir1, testDir2, testDir3, testDir4, testDir5}

	// add a bunch of dummy directories with test files in each one
	var total int
	for _, td := range testDirs {
		tdirs := MakeTestDirs(t, RandInt(50))
		for _, dir := range tdirs {
			if testFiles, err := MakeTestFiles(t, 5); err == nil {
				total += len(testFiles)
				dir.AddFiles(testFiles)
			} else {
				t.Errorf("[ERROR] unable to make test files: %v", err)
			}
		}
		if err := td.AddSubDirs(tdirs); err != nil {
			t.Errorf("[ERROR] unable to add test directories %v", err)
		}
	}

	// make layered file system
	testDir5.AddSubDir(testDir4)
	testDir4.AddSubDir(testDir3)
	testDir3.AddSubDir(testDir2)
	testDir2.AddSubDir(testDir1)

	// run WalkS to get a SyncIndex pointer
	index := testDir5.WalkS()

	assert.NotEqual(t, nil, index)
	assert.NotEqual(t, 0, len(index.LastSync))

	// clean up after testing
	testDir5.Clean(testingDir)
}

func TestWalkF(t *testing.T) {}
