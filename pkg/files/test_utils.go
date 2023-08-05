package files

import (
	"log"
	"os"
	"path/filepath"
	"testing"
)

// run an individual test as part of a series of larger tests
func RunTestStage(t *testing.T, stageName string, test func(t *testing.T)) {
	log.Printf("================================================")
	log.Printf("Running test stage %s\n", stageName)
	test(t)
	log.Printf("================================================")
}

// build path to test file directory
func GetTestingDir() string {
	curDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("[ERROR] unable to get testing directory: %v", err)
	}
	return filepath.Join(curDir, "test_files")
}
