package db

import (
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestAddFile(t *testing.T) {}

func TestAddDirectory(t *testing.T) {}

func TestAddDrive(t *testing.T) {}

func TestAddUser(t *testing.T) {
	testDir := GetTestingDir(t)

	// make testing objects
	_, _, tmpUser := MakeTestItems(t, testDir)
	New(filepath.Join(testDir, "tmp"), CreateUserTable)

	// ------- try to make simple queries to each
	q := NewQuery(filepath.Join(testDir, "tmp"))

	// add test user
	if err := q.AddUser(tmpUser); err != nil {
		t.Fatalf("[ERROR] failed to add user: %v", err)
	}

	// query for user we just added
	u, err := q.GetUser(tmpUser.ID)
	if err != nil {
		t.Fatalf("[ERROR] failed to get user: %v", err)
	}

	assert.Equal(t, tmpUser.ID, u.ID)

	// clean up
	Clean(t, testDir)
}
