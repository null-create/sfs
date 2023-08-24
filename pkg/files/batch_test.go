package files

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestBatchAddFiles(t *testing.T) {
	totalFiles := RandInt(50)
	testFiles, err := MakeABunchOfTxtFiles(totalFiles)
	if err != nil {
		Fatal(t, err)
	}

	// get total size of testFiles in testing dir
	b := NewBatch()
	b.Cap = 1000 // 1 kb batch capacity for testing

	var totalSize int64
	for _, testFile := range testFiles {
		totalSize += testFile.Size()
	}
	assert.True(t, totalSize > b.Cap)

	Clean(t, GetTestingDir())
}

func TestBatchLimit(t *testing.T) {}
