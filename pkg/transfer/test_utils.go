package transfer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	svc "github.com/sfs/pkg/service"
)

const txtData string = "all work and no play makes jack a dull boy\n"

// ---- global

// build path to test file directory. creates testing directory if it doesn't exist.
func GetTestingDir() string {
	curDir, err := os.Getwd()
	if err != nil {
		log.Printf("[ERROR] unable to get testing directory: %v\ncreating...", err)
		if err := os.Mkdir(filepath.Join(curDir, "testing"), 0666); err != nil {
			log.Fatalf("[ERROR] unable to create test directory: %v", err)
		}
	}
	return filepath.Join(curDir, "testing")
}

// handle test failures
//
// like Fatal but you can specify the directory to clean
func Fail(t *testing.T, dir string, err error) {
	if err2 := Clean(t, dir); err2 != nil {
		log.Fatalf("failed to clean testing dir: %v", err)
	}
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
		t.Errorf("[ERROR] unable to read directory: %v", err)
	}

	for _, name := range names {
		if err = os.RemoveAll(filepath.Join(dir, name)); err != nil {
			t.Errorf("[ERROR] unable to remove file: %v", err)
		}
	}

	return nil
}

// ---- files

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
	f := svc.NewFile(filepath.Base(filePath), "me", filePath)
	for i := 0; i < textReps; i++ {
		data += txtData
	}
	if err = f.Save([]byte(data)); err != nil {
		return nil, err
	}
	return f, nil
}

// creates a directory with 10 test files under the testing directory.
func MakeTmpTxtFiles(t *testing.T, path string) (*svc.Directory, error) {
	for i := 0; i < 10; i++ {
		filePath := filepath.Join(path, fmt.Sprintf("tmp-%d.txt", i+1))
		if _, err := MakeTmpTxtFile(filePath, RandInt(1000)); err != nil {
			return nil, err
		}
	}
	dir := svc.NewDirectory("tmp", "some-rand-id", "some-rand-id", path)
	return dir, nil
}
