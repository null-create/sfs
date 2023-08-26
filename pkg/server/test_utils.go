package server

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/sfs/pkg/files"
)

// at ~1 byte per character, and at 49 characters (inlcuding spaces),
// this string is roughly 49 bytes in size, depending on encoding.
//
// in go's case, this is a utf-8 encoded string, so this is roughly 49 bytes
const txtData string = "all work and no play makes jack a dull boy"

// run an individual test as part of a series of larger tests
func RunTestStage(stageName string, test func()) {
	log.Print(fmt.Sprintf("=============== %s Testing Stage ===============", stageName))
	test()
	log.Print("================================================")
}

// build path to test file directory. creates testing directory if it doesn't exist.
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

func GetStateDir() string {
	curDir, err := os.Getwd()
	if err != nil {
		log.Printf("[WARNING] unable to find state file testing directory: %v\ncreating...", err)
		if err := os.Mkdir(filepath.Join(curDir, "testing"), 0666); err != nil {
			log.Fatalf("[ERROR] unable to create state file testing directoryy: %v", err)
		}
	}
	return filepath.Join(curDir, "state")
}

// handle test failures
//
// calls Clean() followed by t.Fatalf()
func Fatal(t *testing.T, err error) {
	Clean(t, GetTestingDir())
	t.Fatalf("[ERROR] %v", err)
}

// make a temp .txt file of n size (in bytes).
//
// n is determined by textReps since that will be how
// many times testData is written to the text file
//
// returns a file pointer to the new temp file
func MakeTmpTxtFile(filePath string, textReps int) (*files.File, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	f := files.NewFile(filepath.Base(filePath), "me", filePath)
	for i := 0; i < textReps; i++ {
		_, err := file.WriteString(txtData)
		if err != nil {
			return nil, fmt.Errorf("error writing to test file: %v", err)
		}
	}
	return f, nil
}

// make a bunch of temp .txt files of varying sizes.
// under pkg/files/testing/tmp
func MakeABunchOfTxtFiles(total int) ([]*files.File, error) {
	tmpDir := filepath.Join(GetTestingDir(), "tmp")

	files := make([]*files.File, 0)
	for i := 0; i < total; i++ {
		fileName := fmt.Sprintf("tmp-%d.txt", i)
		filePath := filepath.Join(tmpDir, fileName)
		f, err := MakeTmpTxtFile(filePath, RandInt(1000))
		if err != nil {
			return nil, fmt.Errorf("error creating temporary file: %v", err)
		}

		files = append(files, f)
	}
	return files, nil
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
