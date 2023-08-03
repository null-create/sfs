package files

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
)

const (
	testData  = "hello, I love you won't you tell me your name?"
	testData2 = "hello, I love you, let me jump in your game"
)

//---------- test fixtures --------------------------------

// makes temp files and file objects for testing purposes
func MakeTestFiles(t *testing.T, total int) ([]*File, error) {
	// build path to test file directory
	curDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("[ERROR] unable to get current working directory: %v", err)
	}
	testDir := filepath.Join(curDir, "test_files")

	// build dummy file objects + test files
	testFiles := make([]*File, 0)
	for i := 0; i < total; i++ {
		name := fmt.Sprintf("%s/test-%d.txt", testDir, i)

		file, err := os.Create(name)
		if err != nil {
			t.Fatalf("[ERROR] failed to create test file: %v", err)
		}
		file.Write([]byte(testData))
		file.Close()

		testFiles = append(testFiles, NewFile(name, "me", name))
	}
	return testFiles, nil
}

func RemoveTestFiles(t *testing.T, total int) error {
	// build path to test file directory
	curDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("[ERROR] unable to get current working directory: %v", err)
	}
	testDir := filepath.Join(curDir, "test_files")

	for i := 0; i < total; i++ {
		name := fmt.Sprintf("%s/test-%d.txt", testDir, i)
		if err := os.Remove(name); err != nil {
			return fmt.Errorf("[ERROR] unable to remove test file: %v", err)
		}
	}
	return nil
}

// ----------------------------------------------------------------

func TestFileIO(t *testing.T) {
	total := RandInt(5)
	testFiles, err := MakeTestFiles(t, total)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// test f.Load()
	for _, f := range testFiles {
		f.Load()
		assert.NotEqual(t, 0, len(f.Content))
		assert.Equal(t, []byte(testData), f.Content)

		f.Clear()
		assert.Equal(t, 0, len(f.Content))
	}

	// update files with new content
	for _, f := range testFiles {
		if err := f.Save([]byte(testData2)); err != nil {
			t.Fatalf("[ERROR] failed to save new content: %v", err)
		}
		f.Load() // f.Save() doesn't load file contents into memory
		assert.NotEqual(t, 0, len(f.Content))
		assert.Equal(t, []byte(testData2), f.Content)

		f.Clear()
		assert.Equal(t, 0, len(f.Content))
	}

	if err := RemoveTestFiles(t, total); err != nil {
		t.Fatalf("[ERROR] failed to remove test files: %v", err)
	}
}

func TestGetFileSize(t *testing.T) {}

func TestFileSecurityFeatures(t *testing.T) {}

func TestFileChecksum(t *testing.T) {}
