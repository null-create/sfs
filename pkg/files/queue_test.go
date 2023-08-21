package files

import (
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
