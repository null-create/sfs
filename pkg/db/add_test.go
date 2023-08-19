package db

import (
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/sfs/pkg/files"
)

func TestAddFile(t *testing.T) {
	testDir := GetTestingDir(t)

	// test db and query
	NewTable(filepath.Join(testDir, "tmp"), CreateFileTable)
	q := NewQuery(filepath.Join(testDir, "tmp"))

	tmpFile := files.NewFile("temp.txt", "bill", filepath.Join(testDir, "tmp"))

	// add temp file
	if err := q.AddFile(tmpFile); err != nil {
		Clean(t, testDir)
		t.Fatalf("[ERROR] failed to add file: %v", err)
	}

	// search for temp file & verify ID
	f, err := q.GetFile(tmpFile.ID)
	if err != nil {
		Clean(t, testDir)
		t.Fatalf("[ERROR] failed to retrieve file: %v", err)
	}
	assert.Equal(t, tmpFile.ID, f.ID)

	// clean up
	Clean(t, testDir)
}

func TestAddDirectory(t *testing.T) {}

func TestAddDrive(t *testing.T) {}

func TestAddUser(t *testing.T) {
	testDir := GetTestingDir(t)

	// make testing objects
	_, _, tmpUser := MakeTestItems(t, testDir)
	NewTable(filepath.Join(testDir, "tmp"), CreateUserTable)

	// test query
	q := NewQuery(filepath.Join(testDir, "tmp"))

	// add test user
	if err := q.AddUser(tmpUser); err != nil {
		Clean(t, testDir)
		t.Fatalf("[ERROR] failed to add user: %v", err)
	}

	// query for user we just added
	u, err := q.GetUser(tmpUser.ID)
	if err != nil {
		Clean(t, testDir)
		t.Fatalf("[ERROR] failed to get user: %v", err)
	}

	assert.Equal(t, tmpUser.ID, u.ID)

	// clean up
	Clean(t, testDir)
}
