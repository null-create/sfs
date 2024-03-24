package service

import (
	"testing"

	"github.com/sfs/pkg/env"

	"github.com/alecthomas/assert/v2"
)

// NOTE: test is flaky
func TestToUpdate(t *testing.T) {
	env.SetEnv(false)

	d := MakeTmpDirs(t)

	// get initial index
	idx := d.WalkS(NewSyncIndex("me"))
	assert.NotEqual(t, nil, idx)
	assert.NotEqual(t, 0, len(idx.LastSync))
	assert.Equal(t, 20, len(idx.LastSync))

	// randomly update some of the files with additional content, causing their
	// last sync times to be updated
	files := d.GetFileMap() // NOTE: this only returns d's top level files
	MutateFiles(t, files)

	// check new index, make sure some of the times are different
	toUpdate := BuildToUpdate(d, idx)
	assert.NotEqual(t, nil, toUpdate)
	assert.NotEqual(t, 0, len(toUpdate.FilesToUpdate))

	// make sure all new sync times are valid
	for _, f := range toUpdate.FilesToUpdate {
		assert.NotEqual(t, 0, f.LastSync.Second())
	}

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Fatal(err)
	}
}

func TestCompare(t *testing.T) {
	env.SetEnv(false)

	// make a temp drive and generate a starting sync index
	tmpDrv := MakeTmpDrive(t)
	origIdx := BuildRootSyncIndex(tmpDrv.Root) // ToUpdate is empty

	// change some files (last three w/e) and generate a new index
	testFiles := tmpDrv.Root.GetFiles()
	victims := testFiles[:3]
	for _, victim := range victims {
		MutateFile(t, victim)
	}
	origIdx = BuildToUpdate(tmpDrv.Root, origIdx)

	// change some more stuff, but update newIdx instead
	newIdx := BuildRootSyncIndex(tmpDrv.Root)
	for _, victim := range victims {
		MutateFile(t, victim)
	}
	newIdx = BuildToUpdate(tmpDrv.Root, newIdx)

	// compare the two indicies
	diffs := Compare(origIdx, newIdx)

	// pre-emptively clean up before the asserts
	if err := Clean(t, GetTestingDir()); err != nil {
		t.Fatal(err)
	}

	// compare
	assert.NotEqual(t, 0, len(diffs.LastSync))
}
