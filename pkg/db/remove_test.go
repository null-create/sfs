package db

import (
	"fmt"
	"log"
	"path/filepath"
	"testing"

	"github.com/sfs/pkg/auth"
	svc "github.com/sfs/pkg/service"
)

func TestAddAndRemoveFile(t *testing.T) {
	testDir := GetTestingDir()

	NewTable(filepath.Join(testDir, "Files"), CreateFileTable)
	q := NewQuery(filepath.Join(testDir, "Files"), false)
	q.Debug = true

	tmpFile := svc.NewFile("temp.txt", "bill", filepath.Join(testDir, "Files"))

	if err := q.AddFile(tmpFile); err != nil {
		Fatal(t, fmt.Errorf("failed to add file: %v", err))
	}
	log.Printf("added file: %s", tmpFile.ID)

	if err := q.RemoveFile(tmpFile.ID); err != nil {
		Fatal(t, fmt.Errorf("failed to remove file: %v", err))
	}
	log.Printf("removed file: %s", tmpFile.ID)

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestAddAndRemoveDir(t *testing.T) {
	testDir := GetTestingDir()

	NewTable(filepath.Join(testDir, "Directories"), CreateDirectoryTable)
	q := NewQuery(filepath.Join(testDir, "Directories"), false)
	q.Debug = true

	tmpDir := svc.NewDirectory("temp.txt", "bill", filepath.Join(testDir, "Directories"))

	if err := q.AddDir(tmpDir); err != nil {
		Fatal(t, fmt.Errorf("failed to add directory: %v", err))
	}
	log.Printf("added directory: %s", tmpDir.ID)

	if err := q.RemoveDirectory(tmpDir.ID); err != nil {
		Fatal(t, fmt.Errorf("failed to remove directory: %v", err))
	}
	log.Printf("removed directory: %s", tmpDir.ID)

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestAddAndRemoveDrive(t *testing.T) {
	testDir := GetTestingDir()

	NewTable(filepath.Join(testDir, "Drives"), CreateDriveTable)
	q := NewQuery(filepath.Join(testDir, "Drives"), false)
	q.Debug = true

	tmpDrive, _, _ := MakeTestItems(t, GetTestingDir())

	if err := q.AddDrive(tmpDrive); err != nil {
		Fatal(t, err)
	}
	log.Printf("added drive: %s", tmpDrive.ID)

	if err := q.RemoveDrive(tmpDrive.ID); err != nil {
		Fatal(t, err)
	}
	log.Printf("removed drive: %s", tmpDrive.ID)

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestAddAndRemoveUser(t *testing.T) {
	testDir := GetTestingDir()

	// test db and query
	NewTable(filepath.Join(testDir, "Users"), CreateUserTable)
	q := NewQuery(filepath.Join(testDir, "Users"), false)
	q.Debug = true

	tmpUser := auth.NewUser("bill", "bill", "bill@bill.com", testDir, false)

	if err := q.AddUser(tmpUser); err != nil {
		Fatal(t, fmt.Errorf("failed to add user: %v", err))
	}
	log.Printf("added user: %s", tmpUser.ID)

	if err := q.RemoveUser(tmpUser.ID); err != nil {
		Fatal(t, fmt.Errorf("failed to remove user: %v", err))
	}
	log.Printf("removed user: %s", tmpUser.ID)

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}

func TestClearTable(t *testing.T) {
	testDir := GetTestingDir()

	// test db and query
	NewTable(filepath.Join(testDir, "Users"), CreateUserTable)
	q := NewQuery(filepath.Join(testDir, "Users"), false)
	q.Debug = true
	q.DBPath = filepath.Join(testDir, "Users")

	tmpUser := auth.NewUser("bill", "bill", "bill@bill.com", testDir, false)

	if err := q.AddUser(tmpUser); err != nil {
		Fatal(t, fmt.Errorf("failed to add user: %v", err))
	}
	log.Printf("[TEST] added user: %s", tmpUser.ID)

	if err := q.ClearTable("users"); err != nil {
		Fatal(t, fmt.Errorf("failed to clear table: %v", err))
	}

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Errorf("[ERROR] unable to remove test directories: %v", err)
	}
}
