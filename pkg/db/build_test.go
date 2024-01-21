package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestBuildDbs(t *testing.T) {
	testDir := GetTestingDir()

	// create tmp databases
	NewTable(filepath.Join(testDir, "users"), CreateUserTable)
	NewTable(filepath.Join(testDir, "files"), CreateFileTable)
	NewTable(filepath.Join(testDir, "drives"), CreateDriveTable)
	NewTable(filepath.Join(testDir, "directories"), CreateDirectoryTable)

	// make sure we created the database files
	entries, err := os.ReadDir(testDir)
	if err != nil {
		Fatal(t, fmt.Errorf("failed to read testing directory: %v", err))
	}
	assert.NotEqual(t, 0, len(entries))

	// get a list of created tables
	files := []string{}
	for _, e := range entries {
		files = append(files, e.Name())
	}

	// check that each DB is present.
	dbs := []string{"users", "files", "drives", "directories"}
	if !AreEqualSlices(files, dbs) {
		Fatal(t, fmt.Errorf("missing entries"))
	}

	// clean up temporary databases
	if err := Clean(t, testDir); err != nil {
		log.Fatal(err)
	}
}
