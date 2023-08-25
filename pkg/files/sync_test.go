package files

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestMakeSyncIndex(t *testing.T) {
	tmpRoot := MakeTmpDirs(t)

	// create a sync index from each test file
	idx := tmpRoot.WalkS()
	assert.NotEqual(t, nil, idx)

	Clean(t, GetTestingDir())
}

// func TestBuildToUpdate(t *testing.T) {

// }
