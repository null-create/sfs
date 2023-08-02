package files

import (
	"fmt"
)

// Queue represents a set of file batches that are to be uploaded or downloaded
type Queue struct {
	Total int `json:"total"`

	Queue []*Batch
}

func NewQ(total int) *Queue {
	return &Queue{
		Total: total,
	}
}

// NOTE: thes does NOT ensure there are no duplicate batches!
// this will need to be done elsewhere
func (q *Queue) Enqueue(b *Batch) {
	q.Queue = append(q.Queue, b)
	q.Total += len(b.Files)
}

func (q *Queue) Dequeue() (*Batch, error) {
	if len(q.Queue) == 0 {
		return nil, fmt.Errorf("[ERROR] Queue is empty")
	}

	item := q.Queue[0]
	q.Queue = q.Queue[1:]
	q.Total -= len(item.Files)

	return item, nil
}
