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
	fileID, err := q.GetFileIDFromPath(tmpFile.Path)
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

func TestFileByUserID(t *testing.T) {
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

	// retrieves all files with this user ID
	files, err := q.GetUsersFiles(tmpFile.OwnerID)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}
	if len(files) == 0 {
		Fail(t, GetTestingDir(), fmt.Errorf("no files found with userID: '%s'", tmpFile.OwnerID))
	}
	for _, file := range files {
		if file.OwnerID != tmpFile.OwnerID {
			Fail(t, GetTestingDir(), fmt.Errorf("db file owner ID does not match original file owner ID"))
		}
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

	// add dir
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

	// add dir
	if err := q.AddDir(tmpDir); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// search by path
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

func TestFindDriveByUserID(t *testing.T) {
	env.SetEnv(false)

	testDir := GetTestingDir()

	// make testing objects
	tmpDrv, _, _ := MakeTestItems(t, GetTestingDir())

	// create tmp table
	NewTable(filepath.Join(testDir, "tmp-db"), CreateDriveTable)

	// test query
	q := NewQuery(filepath.Join(testDir, "tmp-db"), false)

	// add drive
	if err := q.AddDrive(tmpDrv); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// search by userID
	drv, err := q.GetDriveByUserID(tmpDrv.OwnerID)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}
	if drv == nil {
		Fail(t, GetTestingDir(), fmt.Errorf("drive was not found"))
	}
	if drv.OwnerID != tmpDrv.OwnerID {
		Fail(t, GetTestingDir(), fmt.Errorf("owner ID mismatch. orig: %s new: %s", tmpDrv.OwnerID, drv.OwnerID))
	}

	// clean up tmp db
	if err := Clean(t, GetTestingDir()); err != nil {
		log.Fatal(err)
	}
}

func TestFindUsersIdWithDriveID(t *testing.T) {
	env.SetEnv(false)

	testDir := GetTestingDir()

	// make testing objects
	tmpDrv, _, _ := MakeTestItems(t, GetTestingDir())

	// create tmp table
	NewTable(filepath.Join(testDir, "tmp-db"), CreateDriveTable)

	// test query
	q := NewQuery(filepath.Join(testDir, "tmp-db"), false)

	// add user
	if err := q.AddDrive(tmpDrv); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	// search by drive ID
	userID, err := q.GetUserIDFromDriveID(tmpDrv.ID)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}
	if userID == "" {
		Fail(t, GetTestingDir(), fmt.Errorf("userID not found"))
	}

	// clean up tmp db
	if err := Clean(t, GetTestingDir()); err != nil {
		log.Fatal(err)
	}

}
