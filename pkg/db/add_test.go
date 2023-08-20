package db

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/sfs/pkg/files"
)

func TestAddAndFindFile(t *testing.T) {
	testDir := GetTestingDir()

	// test db and query
	NewTable(filepath.Join(testDir, "tmp"), CreateFileTable)
	q := NewQuery(filepath.Join(testDir, "tmp"))

	tmpFile := files.NewFile("temp.txt", "bill", filepath.Join(testDir, "tmp"))

	// add temp file
	if err := q.AddFile(tmpFile); err != nil {
		Fatal(t, fmt.Errorf("failed to add file: %v", err))
	}

	// search for temp file & verify ID
	f, err := q.GetFile(tmpFile.ID)
	if err != nil {
		Fatal(t, fmt.Errorf("failed to get file: %v", err))
	}
	assert.Equal(t, tmpFile.ID, f.ID)

	// clean up
	Clean(t, testDir)
}

func TestAddAndFindDirectory(t *testing.T) {
	testDir := GetTestingDir()

	// test db and query
	NewTable(filepath.Join(testDir, "directories"), CreateDirectoryTable)
	q := NewQuery(filepath.Join(testDir, "directories"))

	_, tmpDir, _ := MakeTestItems(t, GetTestingDir())

	// add tmp directory
	if err := q.AddDir(tmpDir); err != nil {
		Fatal(t, fmt.Errorf("failed to add dir: %v", err))
	}

	// search for temp dir to ensure it was added correctly
	d, err := q.GetDirectory(tmpDir.ID)
	if err != nil {
		Fatal(t, fmt.Errorf("failed to get directory: %v", err))
	}
	// NOTE: d being nil isn't necessarily a problem. if we have a
	// functional table and the directory entry simply doesn't exist,
	// then its not necearily a fault of the program.
	// here we just want to test for the existence of a file so we
	// can ensure the database is working properly.
	assert.NotEqual(t, nil, d)
	assert.Equal(t, tmpDir.ID, d.ID)

	// clean up temporary dir
	Clean(t, GetTestingDir())
}

func TestAddAndFindDrive(t *testing.T) {}

func TestAddAndFindUser(t *testing.T) {
	testDir := GetTestingDir()

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
