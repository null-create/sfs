package files

import (
	"log"
)

// max batch size
//
// (1,000,000,000 / 2^30 bytes/ 1 Gb)
const MAX int64 = 1e+9

// batch represents a collection of files to be uploaded or downloaded
// Batch.limit is set by the network profiler
type Batch struct {
	ID  string // batch ID (UUID)
	Cap int64  // remaining capacity (in bytes)

	Files []*File // files to be uploaded or downloaded
}

func NewBatch() *Batch {
	return &Batch{
		ID:    NewUUID(),
		Cap:   MAX,
		Files: make([]*File, 0),
	}
}

/*
we're bound  by an upper size limit on our batch sizes (MAX) since
we ideally don't want to clog a home network's resources when uploading
or downloading batches of files. MAX is subject to change of course,
but its in place as a mechanism for resource management.

TODO: look at the knapsack problem for guidance here.

NOTE: this does not check whether there are any duplicate files in the
files []*File slice... probably should do that

NOTE: if each batch's files are to be uploaded in their own separate goroutines, and
each thread is part of a waitgroup, and each batch is processed one at a time, then
the runtime for each batch will be bound by the largest file in the batch
(assuming consistent upload speed and no other external circumstances).
*/
func (b *Batch) AddFiles(files []*File) []*File {
	// remember which ones we added so we don't have to modify the
	// files slice in-place as we're iterating over it
	added := make([]*File, 0)
	// remember which files that were passed over for this batch
	notAdded := make([]*File, 0)

	for _, f := range files {
		// "if a current file's size doesn't cause us to exceed the remaining batch capacity, add it."
		//
		// this is basically a greedy approach, but that may change.
		//
		// since lists are unsorted, a file that is much larger than its neighbors may cause
		// batches to not contain as many possible files since one files weight may greatly tip
		// the scales, as it were. NP problems are hard.
		//
		// pre-sorting the list of files will introduce a lower O(nlogn) bound on any possible
		// resulting solution, so our current approach, roughly O(nk) (i think), where n is the
		// number of times we need to iterate over the list of files (and remaning subsets after
		// each batch) and where k is the size of the *current* list we're iterating over and
		// building a batch from (assuming slice shrinkage with each pass).
		//
		// TODO: investigate this case
		// 	- individual files that exceed b.Cap won't be added to the batch ever.
		if b.Cap-f.Size() >= 0 {
			b.Files = append(b.Files, f)
			b.Cap -= f.Size()        // decrement remaning file capacity
			added = append(added, f) // save to added files list
			if b.Cap == 0 {          // don't bother checking the rest
				break
			}
		} else {
			// we want to check the other files in this list
			// since they may be small enough to add onto this batch.
			log.Printf("[DEBUG] file size (%d bytes) exceeds remaining batch capacity (%d bytes).\nattempting to add others...\n", f.Size(), b.Cap)
			notAdded = append(notAdded, f)
			continue
		}
	}

	// TODO: figure out how to communicate which of these scenarious
	// was the case to the caller of this function.

	// success
	if len(added) == len(files) {
		// if all files were successfully added
		log.Printf("[DEBUG] all files added to batch. remaining batch capacity (in bytes): %d", b.Cap)
		return added
	}
	// if we reach capacity before we finish with files,
	// return a list of the remaining files
	if b.Cap == MAX && len(added) < len(files) {
		log.Printf("[DEBUG] reached capacity before we could finish with the remaining files. \nreturning remaining files\n")
		return Diff(added, files)
	}
	// if b.Cap < MAX and we have left over files that were passed over for
	// being to large for the current batch.
	// use AddFiles() again over notAdded list until no more files remain.
	if len(notAdded) > 0 && b.Cap < MAX {
		log.Printf("[DEBUG] returning files passed over for being too large for this batch")
		if len(added) == 0 {
			log.Printf("[WARNING] no files were added!")
		}
		return notAdded
	}
	return nil
}
