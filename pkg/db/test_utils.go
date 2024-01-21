package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/sfs/pkg/auth"
	svc "github.com/sfs/pkg/service"
)

const txtData string = "all work and no play makes jack a dull boy\n"

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

// make a temp .txt file of n size (in bytes).
//
// n is determined by textReps since that will be how
// many times testData is written to the text file
//
// returns a file pointer to the new temp file
func MakeTmpTxtFile(filePath string, textReps int) (*svc.File, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	var data string
	f := svc.NewFile(filepath.Base(filePath), "some-rand-id", "me", filePath)
	for i := 0; i < textReps; i++ {
		data += txtData
	}
	if err = f.Save([]byte(data)); err != nil {
		return nil, err
	}
	return f, nil
}

// make an empty test dir object. does not create a physical directory.
func MakeTestDir(path string) *svc.Directory {
	return svc.NewDirectory("bill", "bill buttlicker", "some-rand-id", filepath.Join(path, "bill"))
}

// make a tmp user for testing
func MakeTestUser(testDir string) *auth.User {
	return auth.NewUser("bill buttlicker", "billb", "bill@bill.com", filepath.Join(testDir, "bill"), false)
}

func MakeTestItems(t *testing.T, testDir string) (*svc.Drive, *svc.Directory, *auth.User) {
	tempDir := svc.NewDirectory(
		"bill", "bill buttlicker", "some-rand-id", filepath.Join(testDir, "bill"),
	)
	tempDrive := svc.NewDrive(
		auth.NewUUID(), "bill", auth.NewUUID(), filepath.Join(testDir, "bill"), tempDir.ID, tempDir,
	)
	tmpUser := auth.NewUser("bill", "bill123", "bill@bill.com", testDir, false)
	return tempDrive, tempDir, tmpUser
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
