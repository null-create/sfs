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
	q := NewQuery(filepath.Join(testDir, "Files"), false)
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

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestAddAndFindMultipleFiles(t *testing.T) {
	testDir := GetTestingDir()

	// test db and query
	NewTable(filepath.Join(testDir, "Files"), CreateFileTable)
	q := NewQuery(filepath.Join(testDir, "Files"), false)
	q.Debug = true

	// make a bunch of dummy files
	total := files.RandInt(100)
	testFiles := make([]*files.File, 0)
	for i := 0; i < total; i++ {
		fn := fmt.Sprintf("test-%d.txt", i)
		f := files.NewFile(fn, "me", filepath.Join(testDir, fn))
		testFiles = append(testFiles, f)
	}

	// add files to db
	if err := q.AddFiles(testFiles); err != nil {
		Fatal(t, err)
	}

	// attempt to retrieve all the test files we just added
	results, err := q.GetFiles()
	if err != nil {
		Fatal(t, err)
	}
	assert.NotEqual(t, nil, results)
	assert.Equal(t, total, len(results))
	assert.Equal(t, len(testFiles), len(results))

	// TODO: rework this to skip lastSync as it's updated
	// when we pull the data from the db

	// for i, testFile := range results {
	// 	assert.Equal(t, testFiles[i], testFile)
	// }

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestAddAndFindDirectory(t *testing.T) {
	testDir := GetTestingDir()

	// test db and query
	NewTable(filepath.Join(testDir, "directories"), CreateDirectoryTable)
	q := NewQuery(filepath.Join(testDir, "directories"), false)

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

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestAddAndFindDrive(t *testing.T) {
	testDir := GetTestingDir()

	// make testing objects
	tmpDrive, _, _ := MakeTestItems(t, testDir)
	NewTable(filepath.Join(testDir, "tmp"), CreateDriveTable)

	// test query
	q := NewQuery(filepath.Join(testDir, "tmp"), false)
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

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestAddAndFindUser(t *testing.T) {
	testDir := GetTestingDir()

	// make testing objects
	_, _, tmpUser := MakeTestItems(t, testDir)
	NewTable(filepath.Join(testDir, "tmp"), CreateUserTable)

	// test query
	q := NewQuery(filepath.Join(testDir, "tmp"), false)

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

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

// func TestUserExists(t *testing.T) {}
