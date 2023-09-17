package service

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestToUpdate(t *testing.T) {
	d := MakeTmpDirs(t)

	// get initial index
	idx := d.WalkS()
	assert.NotEqual(t, nil, idx)
	assert.NotEqual(t, 0, len(idx.LastSync))
	assert.Equal(t, 20, len(idx.LastSync))

	// randomly update some of the files with additional content, causing their
	// last sync times to be updated
	files := d.GetFiles()
	for _, f := range files {
		if RandInt(2) == 1 {
			if err := f.Save([]byte(testData)); err != nil {
				Fatal(t, err)
			}
		}
	}

	// check new index, make sure some of the times are different
	toUpdate := BuildToUpdate(d, idx)
	assert.NotEqual(t, nil, toUpdate)
	assert.NotEqual(t, 0, len(toUpdate.LastSync))

	// make sure all new sync times are valid
	for _, f := range toUpdate.ToUpdate {
		assert.NotEqual(t, 0, f.LastSync.Second())
	}

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Fatal(err)
	}
}
