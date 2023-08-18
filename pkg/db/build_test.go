package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestBuildDbs(t *testing.T) {
	testDir := GetTestingDir(t)

	// ------- create tmp databases
	NewDB("users", filepath.Join(testDir, "users"))
	NewDB("files", filepath.Join(testDir, "files"))
	NewDB("drives", filepath.Join(testDir, "drives"))
	NewDB("directories", filepath.Join(testDir, "directories"))

	// make sure we created the database files
	entries, err := os.ReadDir(testDir)
	if err != nil {
		t.Fatalf("[ERROR] failed to read testing directory: %v", err)
	}
	assert.NotEqual(t, 0, len(entries))

	// ------- clean up temporary databases
	Clean(t, testDir)
}
