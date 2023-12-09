package service

import (
	"path/filepath"
	"testing"

	"github.com/sfs/pkg/env"

	"github.com/alecthomas/assert/v2"
)

func TestQueueOrder(t *testing.T) {
	env.BuildEnv(false)

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
	env.BuildEnv(false)

	_, err := MakeTmpDir(t, filepath.Join(GetTestingDir(), "tmp"))
	if err != nil {
		Fatal(t, err)
	}
	f, err := MakeABunchOfTxtFiles(25)
	if err != nil {
		Fatal(t, err)
	}

	b := NewBatch()
	b.Cap = int64(TEST_MAX)

	q := buildQ(f, b, NewQ())
	assert.NotEqual(t, nil, q)
	assert.NotEqual(t, 0, len(q.Queue))

	// clean up
	if err := Clean(t, GetTestingDir()); err != nil {
		t.Fatal(err)
	}
}

func TestBuildQWithLotsOfDifferentFiles(t *testing.T) {
	env.BuildEnv(false)

	_, err := MakeTmpDir(t, filepath.Join(GetTestingDir(), "tmp"))
	if err != nil {
		Fatal(t, err)
	}
	f, err := MakeABunchOfTxtFiles(50)
	if err != nil {
		Fatal(t, err)
	}

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
	env.BuildEnv(false)

	d, err := MakeTmpDir(t, filepath.Join(GetTestingDir(), "tmp"))
	if err != nil {
		Fatal(t, err)
	}
	f, err := MakeABunchOfTxtFiles(50)
	if err != nil {
		Fatal(t, err)
	}
	d.AddFiles(f)

	b := NewBatch()
	b.Cap = 100 // set a small capacity

	// build a sync index and mutate files so we can
	// add them to the ToUpdate map (BuildQ checks for this)
	idx := BuildSyncIndex(d)
	d.Files = MutateFiles(t, d.Files)
	idx = BuildToUpdate(d, idx)

	// should return a "large file" queue, i.e just a
	// queue of each of the files.
	q := BuildQ(idx)

	assert.NotEqual(t, nil, q)
	assert.NotEqual(t, 0, len(q.Queue))

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Fatal(err)
	}
}
