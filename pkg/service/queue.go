package service

// Queue represents a set of file batches that are to be uploaded or downloaded
type Queue struct {
	Queue []*Batch // a batch represents a collection of files to be uploaded or downloaded
}

// create a new upload/download queue
func NewQ() *Queue {
	return &Queue{
		Queue: make([]*Batch, 0),
	}
}

// NOTE: this does NOT ensure there are no duplicate batches!
//
// that will need to be done elsewhere
func (q *Queue) Enqueue(b *Batch) {
	q.Queue = append(q.Queue, b)
}

func (q *Queue) Dequeue() *Batch {
	if len(q.Queue) == 0 {
		return nil
	}
	item := q.Queue[0]
	q.Queue = q.Queue[1:]
	return item
}
