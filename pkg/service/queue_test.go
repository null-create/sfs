package service

import (
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestQueueOrder(t *testing.T) {
	testQ := NewQ()

	testBatch1 := NewBatch()
	testQ.Enqueue(testBatch1)

	testBatch2 := NewBatch()
	testQ.Enqueue(testBatch2)

	testBatch3 := NewBatch()
	testQ.Enqueue(testBatch3)

	t1, err := testQ.Dequeue()
	if err != nil {
		Fatal(t, err)
	}
	assert.Equal(t, testBatch1.ID, t1.ID)

	t2, err := testQ.Dequeue()
	if err != nil {
		Fatal(t, err)
	}
	assert.Equal(t, testBatch2.ID, t2.ID)

	t3, err := testQ.Dequeue()
	if err != nil {
		Fatal(t, err)
	}
	assert.Equal(t, testBatch3.ID, t3.ID)
}

func TestBuildQueue(t *testing.T) {
	tmpRoot := MakeTmpDirs(t)

	// get initial index
	idx := tmpRoot.WalkS(NewSyncIndex("me"))

	// randomly update some of the files with additional content, causing their
	// last sync times to be updated
	files := tmpRoot.GetFiles()
	MutateFiles(t, files)

	// check new index, make sure some of the times are different
	toUpdate := BuildToUpdate(tmpRoot, idx)

	// build and inspect file queue
	q := BuildQ(toUpdate, NewQ())
	assert.NotEqual(t, nil, q)
	assert.NotEqual(t, 0, len(q.Queue))

	var total int
	for _, batch := range q.Queue {
		total += batch.Total
	}
	assert.Equal(t, total, len(toUpdate.ToUpdate))
	assert.Equal(t, total, q.Queue[0].Total)

	// clean up
	if err := Clean(t, GetTestingDir()); err != nil {
		t.Fatal(err)
	}
}

func TestBuildQWithLotsOfDifferentFiles(t *testing.T) {
	d, err := MakeTmpDir(t, filepath.Join(GetTestingDir(), "tmp"))
	if err != nil {
		Fatal(t, err)
	}
	f, err := MakeABunchOfTxtFiles(50)
	if err != nil {
		Fatal(t, err)
	}

	MutateFiles(t, d.GetFiles())

	b := NewBatch()
	b.Cap = int64(TEST_MAX)

	q := buildQ(f, b, NewQ())
	assert.NotEqual(t, nil, q)
	assert.NotEqual(t, 0, len(q.Queue))
	assert.True(t, len(q.Queue) < len(b.Files))

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Fatal(err)
	}
}

func TestBuildQWithFilesLargerThanMAX(t *testing.T) {
	d, err := MakeTmpDir(t, filepath.Join(GetTestingDir(), "tmp"))
	if err != nil {
		Fatal(t, err)
	}
	f, err := MakeABunchOfTxtFiles(50)
	if err != nil {
		Fatal(t, err)
	}
	d.AddFiles(f)

	idx := BuildSyncIndex(d)
	assert.NotEqual(t, nil, idx)

	MutateFiles(t, d.GetFiles())
	idx = BuildToUpdate(d, idx)

	b := NewBatch()
	b.Cap = 100 // set a small capacity

	// should return a "large file" queue, i.e just a
	// queue of each of the files.
	q := BuildQ(idx, NewQ())
	assert.NotEqual(t, nil, q)
	assert.NotEqual(t, 0, len(q.Queue))

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Fatal(err)
	}
}
