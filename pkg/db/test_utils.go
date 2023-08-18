package db

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/files"
)

// build path to test file directory. creates testing directory if it doesn't exist.
func GetTestingDir(t *testing.T) string {
	curDir, err := os.Getwd()
	if err != nil {
		log.Printf("[ERROR] unable to get testing directory: %v\ncreating...", err)
		if err := os.Mkdir(filepath.Join(curDir, "testing"), 0666); err != nil {
			t.Fatalf("[ERROR] unable to create test directory: %v", err)
		}
	}
	return filepath.Join(curDir, "testing")
}

// clean all contents from the testing directory
func Clean(t *testing.T, dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		t.Errorf("[ERROR] unable to read directory: %v", err)
	}

	for _, name := range names {
		if err = os.RemoveAll(filepath.Join(dir, name)); err != nil {
			t.Errorf("[ERROR] unable to remove file: %v", err)
		}
	}

	return nil
}

func MakeTestItems(t *testing.T, testDir string) (*files.Drive, *files.Directory, *auth.User) {
	tempDir := files.NewDirectory(
		"bill",
		"bill stinkwater",
		filepath.Join(testDir, "bill"),
	)
	tempDrive := files.NewDrive(
		files.NewUUID(),
		"bill",
		"bill stinkwater",
		filepath.Join(testDir, "bill"),
		tempDir,
	)
	tmpUser := auth.NewUser(
		"bill",
		"bill123",
		"bill@bill.com",
		tempDrive,
		false,
	)
	return tempDrive, tempDir, tmpUser
}
