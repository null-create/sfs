package db

import (
	"fmt"
	"log"
	"path/filepath"
	"testing"

	"github.com/sfs/pkg/env"
)

func TestFindFileIdByPath(t *testing.T) {
	env.SetEnv(false)

	testDir := GetTestingDir()

	// make testing objects
	tmpFile, err := MakeTmpTxtFile(filepath.Join(GetTestingDir(), "tmp.txt"), RandInt(1000))
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// create tmp table
	NewTable(filepath.Join(testDir, "tmp-db"), CreateFileTable)

	// test query
	q := NewQuery(filepath.Join(testDir, "tmp-db"), false)

	// add file
	if err := q.AddFile(tmpFile); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// search by file ID
	fileID, err := q.GetFileID(tmpFile.Path)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}
	if fileID == "" {
		Fail(t, GetTestingDir(), fmt.Errorf("missing file ID"))
	}
	if fileID != tmpFile.ID {
		Fail(t, GetTestingDir(), fmt.Errorf("fileID mismatch. orig: %s, new: %s", tmpFile.ID, fileID))
	}

	// clean up tmp db
	if err := Clean(t, GetTestingDir()); err != nil {
		log.Fatal(err)
	}
}

func TestFindFileByName(t *testing.T) {
	env.SetEnv(false)

	testDir := GetTestingDir()

	// make testing objects
	tmpFile, err := MakeTmpTxtFile(filepath.Join(GetTestingDir(), "tmp.txt"), RandInt(1000))
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// create tmp table
	NewTable(filepath.Join(testDir, "tmp-db"), CreateFileTable)

	// test query
	q := NewQuery(filepath.Join(testDir, "tmp-db"), false)

	// add file
	if err := q.AddFile(tmpFile); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// search by file name
	file, err := q.GetFileByName(tmpFile.Name)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}
	if file == nil {
		Fail(t, GetTestingDir(), fmt.Errorf("file not found: %v", tmpFile.Name))
	}
	if file.Name != tmpFile.Name {
		Fail(t, GetTestingDir(), fmt.Errorf("file name mismatch. orig: %s new: %s", tmpFile.Name, file.Name))
	}

	// clean up tmp db
	if err := Clean(t, GetTestingDir()); err != nil {
		log.Fatal(err)
	}
}

func TestFindFileByPath(t *testing.T) {
	env.SetEnv(false)

	testDir := GetTestingDir()

	// make testing objects
	tmpFile, err := MakeTmpTxtFile(filepath.Join(GetTestingDir(), "tmp.txt"), RandInt(1000))
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// create tmp table
	NewTable(filepath.Join(testDir, "tmp-db"), CreateFileTable)

	// test query
	q := NewQuery(filepath.Join(testDir, "tmp-db"), false)

	// add file
	if err := q.AddFile(tmpFile); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// search by file path
	file, err := q.GetFileByPath(tmpFile.Path)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}
	if file == nil {
		Fail(t, GetTestingDir(), fmt.Errorf("file not found: %v", tmpFile.Name))
	}
	if file.Path != tmpFile.Path {
		Fail(t, GetTestingDir(), fmt.Errorf("file name mismatch. orig: %s new: %s", tmpFile.Name, file.Name))
	}

	// clean up tmp db
	if err := Clean(t, GetTestingDir()); err != nil {
		log.Fatal(err)
	}
}

func TestFindDirByName(t *testing.T) {
	env.SetEnv(false)

	testDir := GetTestingDir()

	// make testing objects
	tmpDir := MakeTestDir(filepath.Join(GetTestingDir(), "tmp"))

	// create tmp table
	NewTable(filepath.Join(testDir, "tmp-db"), CreateDirectoryTable)

	// test query
	q := NewQuery(filepath.Join(testDir, "tmp-db"), false)

	// add file
	if err := q.AddDir(tmpDir); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// search by name
	dir, err := q.GetDirectoryByName(tmpDir.Name)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}
	if dir == nil {
		Fail(t, GetTestingDir(), fmt.Errorf("dir not found"))
	}
	if dir.Name != tmpDir.Name {
		Fail(t, GetTestingDir(), fmt.Errorf("dir name mismatch. orig: %s new: %s", tmpDir.Name, dir.Name))
	}

	// clean up tmp db
	if err := Clean(t, GetTestingDir()); err != nil {
		log.Fatal(err)
	}
}

func TestFindDirByPaths(t *testing.T) {
	env.SetEnv(false)

	testDir := GetTestingDir()

	// make testing objects
	tmpDir := MakeTestDir(filepath.Join(GetTestingDir(), "tmp"))

	// create tmp table
	NewTable(filepath.Join(testDir, "tmp-db"), CreateDirectoryTable)

	// test query
	q := NewQuery(filepath.Join(testDir, "tmp-db"), false)

	// add file
	if err := q.AddDir(tmpDir); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// search by name
	dir, err := q.GetDirectoryByPath(tmpDir.Path)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}
	if dir == nil {
		Fail(t, GetTestingDir(), fmt.Errorf("dir not found"))
	}
	if dir.Path != tmpDir.Path {
		Fail(t, GetTestingDir(), fmt.Errorf("dir name mismatch. orig: %s new: %s", tmpDir.Name, dir.Name))
	}

	// clean up tmp db
	if err := Clean(t, GetTestingDir()); err != nil {
		log.Fatal(err)
	}
}

func TestFindDriveByUserID(t *testing.T) {}

func TestFindDriveIDWithUserID(t *testing.T) {}

func TestFindUsersIdWithDriveID(t *testing.T) {}
