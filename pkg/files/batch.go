package files

import (
	"log"
)

// max batch size
//
// (1,000,000,000 / 1,073,741,824 / 2^30 bytes)
const MAX int64 = 1e+9

// batch represents a collection of files to be uploaded or downloaded
// Batch.limit is set by the network profiler
type Batch struct {
	ID  string `json:"id"`  // batch ID (UUID)` // total FILE SIZE in mb allowable for this batch.
	Cap int64  `json:"cap"` // remaining capacity (in bytes)

	// files to be uploaded or downloaded
	Files []*File
}

func NewBatch() *Batch {
	return &Batch{
		ID:  NewUUID(),
		Cap: MAX,
	}
}

// return the difference between two []*File slices.
//
// assuming that go's map implementation has ~O(1) access time,
// then this function should work in ~O(n) on an unsorted slice.
//
// https://stackoverflow.com/questions/19374219/how-to-find-the-difference-between-two-slices-of-strings
func diff(f, g []*File) []*File {
	tmp := make(map[*File]string, len(g))
	for _, file := range g {
		tmp[file] = "hi"
	}
	var diff []*File
	for _, file := range f {
		if _, found := tmp[file]; !found {
			diff = append(diff, file)
		}
	}
	return diff
}

/*
TODO: look at the knapsack problem for guidance here.

we're bound  by an upper size limit on our batch sizes (MAX) since
we ideally don't want to clog a home network's resources when uploading
or downloading batches of files. MAX is subject to change of course,
but it's in place as a mechanism for resource management.
*/

// NOTE this does not check whether there are any duplicate files in the
// files []*File slice... probably should do that
func (b *Batch) AddFiles(files []*File) []*File {
	// remember which ones we added so we don't have to modify files slice in-place
	// as we're iterating over it
	added := make([]*File, 0)
	// remember which files that were too big for this batch
	notAdded := make([]*File, 0)

	for _, f := range files {
		// this is basically a greedy approach, but that may
		// be subect to change.
		if b.Cap-f.Size() >= 0 {
			b.Files = append(b.Files, f)
			b.Cap -= f.Size()        // update remaning file capacity
			added = append(added, f) // save to added files list
		} else {
			// we want to check the other files in this list
			// since they may be small enough to add onto this batch.
			log.Printf("[DEBUG] file size (%d bytes) exceeds remaining batch capacity (%d bytes).\nattempting to add others...\n", f.Size(), b.Cap)
			notAdded = append(notAdded, f)
			continue
		}
	}

	if len(added) == len(files) {
		// if all files were successfully added
		log.Printf("[DEBUG] all files added to batch. remaining batch capacity (in bytes): %d", b.Cap)
		return added
	} else if b.Cap == MAX && len(added) < len(files) {
		// if we reach capacity before we finish with files,
		// return a list of the remaining files
		return diff(added, files)
	} else if len(notAdded) > 0 && b.Cap < MAX {
		// if b.Cap < MAX and we have left over files that were passed over for
		// being to large for the current batch
		return notAdded
	}
	return nil
}
