package db

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestBuildDbs(t *testing.T) {
	testDir := GetTestingDir()

	// ------- create tmp databases
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

	// ------- clean up temporary databases
	Clean(t, testDir)
}
