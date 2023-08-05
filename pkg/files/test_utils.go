package files

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
)

// run an individual test as part of a series of larger tests
func RunTestStage(t *testing.T, stageName string, test func(t *testing.T)) {
	log.Print(fmt.Sprintf("=============== %s Testing Stage ===============", stageName))
	test(t)
	log.Print("================================================")
}

// build path to test file directory. creates testing directory if it doesn't exist.
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
