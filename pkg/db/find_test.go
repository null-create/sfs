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
		if err2 := Clean(t, GetTestingDir()); err2 != nil {
			log.Fatalf("failed to clean tmp dir: " + err2.Error())
		}
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

func TestFindFileByName(t *testing.T) {}

func TestFindUserByID(t *testing.T) {}

func TestFindDirByName(t *testing.T) {}

func TestFindDirByPaths(t *testing.T) {}

func TestFindDriveByUserID(t *testing.T) {}

func TestFindDriveIDWithUserID(t *testing.T) {}

func TestFindUsersIdWithDriveID(t *testing.T) {}
