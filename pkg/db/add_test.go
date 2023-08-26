package db

import (
	"fmt"
	"log"
	"path/filepath"
	"testing"

	"github.com/sfs/pkg/files"

	"github.com/alecthomas/assert/v2"
)

func TestAddAndFindFile(t *testing.T) {
	testDir := GetTestingDir()

	// test db and query
	NewTable(filepath.Join(testDir, "Files"), CreateFileTable)
	q := NewQuery(filepath.Join(testDir, "Files"))
	q.Debug = true

	tmpFile := files.NewFile("temp.txt", "bill", filepath.Join(testDir, "Files"))

	// add temp file
	if err := q.AddFile(tmpFile); err != nil {
		Fatal(t, fmt.Errorf("failed to add file: %v", err))
	}
	log.Printf("added file: %s", tmpFile.ID)

	// search for temp file & verify ID
	f, err := q.GetFile(tmpFile.ID)
	if err != nil {
		Fatal(t, fmt.Errorf("failed to get file: %v", err))
	}
	// NOTE: f being nil isn't necessarily a problem. if we have a
	// functional table and the entry simply doesn't exist,
	// then its not necessarily a failure -- the item may simply not exist.
	// here we just want to test for the existence of a file so we
	// can ensure database I/O is working properly.
	assert.NotEqual(t, nil, f)
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
	assert.NotEqual(t, nil, d)
	assert.Equal(t, tmpDir.ID, d.ID)

	// clean up temporary dir
	Clean(t, GetTestingDir())
}

func TestAddAndFindDrive(t *testing.T) {
	testDir := GetTestingDir()

	// make testing objects
	tmpDrive, _, _ := MakeTestItems(t, testDir)
	NewTable(filepath.Join(testDir, "tmp"), CreateDriveTable)

	// test query
	q := NewQuery(filepath.Join(testDir, "tmp"))
	if err := q.AddDrive(tmpDrive); err != nil {
		Fatal(t, err)
	}

	// add test drive
	d, err := q.GetDrive(tmpDrive.ID)
	if err != nil {
		Fatal(t, err)
	}
	assert.NotEqual(t, nil, d)
	assert.Equal(t, tmpDrive.ID, d.ID)

	Clean(t, GetTestingDir())
}

func TestAddAndFindUser(t *testing.T) {
	testDir := GetTestingDir()

	// make testing objects
	_, _, tmpUser := MakeTestItems(t, testDir)
	NewTable(filepath.Join(testDir, "tmp"), CreateUserTable)

	// test query
	q := NewQuery(filepath.Join(testDir, "tmp"))

	// add test user
	if err := q.AddUser(tmpUser); err != nil {
		Fatal(t, err)
	}

	// query for user we just added
	u, err := q.GetUser(tmpUser.ID)
	if err != nil {
		Fatal(t, err)
	}

	assert.Equal(t, tmpUser.ID, u.ID)

	// clean up
	Clean(t, testDir)
}

// func TestUserExists(t *testing.T) {}
