package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/files"
)

// handles test failures,
// supplies [ERROR] prefix to supplied error messages
//
// it's just a call to Clean() followed by
// a call to t.Fatalf()
func Fatal(t *testing.T, err error) {
	if err2 := Clean(t, GetTestingDir()); err2 != nil {
		log.Printf("[ERROR] failed to clean testing directory during recovery: %v", err2)
	}
	t.Fatalf("[ERROR] %v", err)
}

func GetTestingDir() string {
	curDir, err := os.Getwd()
	if err != nil {
		log.Printf("[WARNING] unable to get testing directory: %v\ncreating...", err)
		if err := os.Mkdir(filepath.Join(curDir, "testing"), 0666); err != nil {
			log.Fatalf("[ERROR] unable to create test directory: %v", err)
		}
	}
	return filepath.Join(curDir, "testing")
}

// clean all contents from the testing directory
func Clean(t *testing.T, dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("[ERROR] could not open test directory: %v", err)
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
