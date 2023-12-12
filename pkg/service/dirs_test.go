package service

import (
	"fmt"
	"log"
	"path/filepath"
	"testing"

	"github.com/sfs/pkg/env"

	"github.com/alecthomas/assert/v2"
)

func TestSecurityFeatures(t *testing.T) {
	env.BuildEnv(false)

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

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestAddFiles(t *testing.T) {
	env.BuildEnv(false)

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

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestRemoveFiles(t *testing.T) {
	env.BuildEnv(false)

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

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestAddSubDir(t *testing.T) {
	env.BuildEnv(false)

	dir1 := NewDirectory("tmp", "me", filepath.Join(GetTestingDir(), "tmp1"))
	dir2 := NewDirectory("tmp2", "me", filepath.Join(GetTestingDir(), "tmp2"))
	if err := dir1.AddSubDir(dir2); err != nil {
		Fatal(t, err)
	}
	assert.NotEqual(t, 0, len(dir1.Dirs))
	assert.True(t, dir1.HasDir(dir2.ID))

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestAddSubDirs(t *testing.T) {
	env.BuildEnv(false)

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
	env.BuildEnv(false)

	total := RandInt(100)
	// root test subdir
	testRoot := NewRootDirectory("tmp", "me", GetTestingDir())
	// subdirs to add
	testDirs := MakeTestDirs(t, total)

	if err := testRoot.AddSubDirs(testDirs); err != nil {
		Fatal(t, err)
	}
	assert.Equal(t, total, len(testRoot.Dirs))

	if err := testRoot.RemoveSubDirs(); err != nil {
		Fatal(t, err)
	}
	assert.NotEqual(t, total, len(testRoot.Dirs))
	assert.Equal(t, 0, len(testRoot.Dirs))
}

func TestGetDirSize(t *testing.T) {
	env.BuildEnv(false)

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

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestWalk(t *testing.T) {
	env.BuildEnv(false)

	tmpRoot := MakeTmpDirs(t)
	items := tmpRoot.Walk()

	files := items.GetFiles()
	if len(files) == 0 {
		Fail(t, GetTestingDir(), fmt.Errorf("no test files found"))
	}

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestWalkF(t *testing.T) {
	env.BuildEnv(false)

	d := MakeTmpDirs(t)
	fileToFind := NewFile("findMe", "bill", filepath.Join(d.Path, "findMe"))
	d.addFile(fileToFind)

	found := d.WalkF(fileToFind.ID)

	assert.NotEqual(t, nil, found)
	assert.Equal(t, fileToFind.ID, found.ID)

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestWalkD(t *testing.T) {
	env.BuildEnv(false)

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
	dir := testDir5.WalkD(idToFind)
	assert.NotEqual(t, nil, dir)
	assert.Equal(t, idToFind, dir.ID)
}

func TestWalkS(t *testing.T) {
	env.BuildEnv(false)

	tmpDir := MakeTmpDirs(t)

	idx := tmpDir.WalkS(NewSyncIndex("me"))

	assert.NotEqual(t, nil, idx)
	assert.NotEqual(t, 0, len(idx.LastSync))
	assert.Equal(t, 20, len(idx.LastSync)) // this is how many test files were generated

	// make sure there's actual times and not uninstantiated time.Time objects
	for _, lastSync := range idx.LastSync {
		assert.NotEqual(t, 0, lastSync.Second())
	}

	// clean up after testing
	if err := Clean(t, GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestWalkU(t *testing.T) {
	env.BuildEnv(false)

	tmpDir := MakeTmpDirs(t)

	idx := tmpDir.WalkS(NewSyncIndex("me"))

	assert.NotEqual(t, nil, idx)
	assert.NotEqual(t, 0, len(idx.LastSync))
	assert.Equal(t, 20, len(idx.LastSync)) // this is how many test files were generated

	// make sure there's actual times and not uninstantiated time.Time objects
	for _, lastSync := range idx.LastSync {
		assert.NotEqual(t, 0, lastSync.Second())
	}

	// TODO: update a few files at random (remember tatalUpdate amount), using
	// their built in f.Save() method.
	// this should update that specific file's LastSync field, and therefore should
	// make it eligible for a sync event.
	// we should check that len(idx.ToUpdate) == tatalUpdate
	// totalUpdate := RandInt(9)

	// var updated int

	// clean up after testing
	if err := Clean(t, GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestWalkO(t *testing.T) {
	env.BuildEnv(false)

	tmpDir := MakeTmpDirs(t)

	// create a test function to pass to WalkF()
	if err := tmpDir.WalkO(func(file *File) error {
		if file == nil {
			return fmt.Errorf("test file pointer is nil")
		}
		log.Printf("[TEST] this is a test file (id=%s)", file.ID)
		return nil
	}); err != nil {
		Fatal(t, err)
	}

	// clean up after testing
	if err := Clean(t, GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestWalkFiles(t *testing.T) {
	env.BuildEnv(false)

	tmpDir := MakeTmpDirs(t)
	files := tmpDir.WalkFs()
	assert.NotEqual(t, nil, files)
	assert.NotEqual(t, 0, len(files))
	assert.Equal(t, 20, len(files))

	// clean up after testing
	if err := Clean(t, GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}
