package db

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/files"
)

func GetTestingDir() string {
	curDir, err := os.Getwd()
	if err != nil {
		log.Printf("[ERROR] unable to get testing directory: %v\ncreating...", err)
		if err := os.Mkdir(filepath.Join(curDir, "test_files"), 0666); err != nil {
			log.Fatalf("[ERROR] unable to create test directory: %v", err)
		}
	}
	return filepath.Join(curDir, "test_files")
}

// handle test failures
func Fatal(t *testing.T, err error) {
	Clean(t, GetTestingDir())
	t.Fatalf("[ERROR] %v", err)
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
		t.Errorf("[ERROR] unable to read directory (%s): \n%v\n", dir, err)
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
