package db

import (
	"fmt"
	"log"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
	svc "github.com/sfs/pkg/service"
)

func TestCreateAndUpdateAFile(t *testing.T) {
	testDir := GetTestingDir()

	// test db and query
	NewTable(filepath.Join(testDir, "Files"), CreateFileTable)
	q := NewQuery(filepath.Join(testDir, "Files"), false)
	q.Debug = true

	tmpFile := svc.NewFile("temp.txt", "bill", filepath.Join(testDir, "Files"))

	// add temp file
	if err := q.AddFile(tmpFile); err != nil {
		t.Fatal(err)
	}

	tmpFile.Name = "some-doc.txt"

	if err := q.UpdateFile(tmpFile); err != nil {
		t.Fatal(err)
	}
	tmp, err := q.GetFile(tmpFile.ID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, tmpFile.Name, tmp.Name)

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Fatal(err)
	}
}

func TestCreateAndUpdateADirectory(t *testing.T) {
	testDir := GetTestingDir()

	// test db and query
	NewTable(filepath.Join(testDir, "Directories"), CreateDirectoryTable)
	q := NewQuery(filepath.Join(testDir, "Directories"), false)
	q.Debug = true

	tmpDir := svc.NewDirectory("tmp", "bill buttlicker", filepath.Join(testDir, "Directories"))

	// add temp directory
	if err := q.AddDir(tmpDir); err != nil {
		t.Fatal(err)
	}
	log.Printf("added directory: %s", tmpDir.ID)

	// update the directory
	tmpDir.Name = "pron"

	if err := q.UpdateDir(tmpDir); err != nil {
		t.Fatal(fmt.Errorf("failed to update directory: %v", err))
	}
	// pull the dir from the db and make sure the name is accurate
	d, err := q.GetDirectory(tmpDir.ID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, tmpDir.Name, d.Name)

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Fatal(err)
	}
}

func TestCreateAndUpdateADrive(t *testing.T) {}

func TestCreateAndUpdateAUser(t *testing.T) {}
