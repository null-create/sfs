package db

import (
	"path/filepath"
	"testing"
)

func TestAddUser(t *testing.T) {
	testDir := GetTestingDir(t)

	// make testing objects
	_, _, tmpUser := MakeTestItems(t, testDir)

	// ------- try to make simple queries to each
	q1 := NewQuery(filepath.Join(testDir, "users"))

	// add test user
	q1.Connect()
	if err := q1.AddUser(tmpUser); err != nil {
		t.Fatalf("[ERROR] failed to add user: %v", err)
	}
	q1.Close()

	// add test drive

	// add test directory

	// query for each item in each db to make sure they were added successfully
}

func TestAddFile(t *testing.T) {}

func TestAddDirectory(t *testing.T) {}

func TestAddDrive(t *testing.T) {}
