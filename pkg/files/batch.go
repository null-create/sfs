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

/*
TODO: look at the knapsack problem for guidance here.

we're bound  by an upper size limit on our batch sizes (MAX) since
we ideally don't want to clog a home network's resources when uploading
or downloading batches of files. MAX is subject to change of course,
but its in place as a mechanism for resource management.
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
		// this is basically a greedy approach, but that may be subect to change.
		//
		// the main criteria is essentially that whatever
		// the size of the *current* file is, it shouldn't cause us to exceed our
		// capacity (b.Cap = MAX), so even if a list of files has an unsorted
		// series of valies, we will only ever care about the effect of the *current* file's size
		// in relation to the remaining capacity of the batch, not the sum total
		// of the given file list (sorted or otherwise).
		//
		// since lists are unsorted, a file that is much larger than its neighbors may cause
		// batches to not contain as many possible files since one files weight may greatly tip the scales,
		// as it were. NP problems are hard.
		//
		// individual files that exceed b.Cap won't be added to the batch ever.
		// TODO: investigate this case
		if b.Cap-f.Size() >= 0 {
			b.Files = append(b.Files, f)
			b.Cap -= f.Size()        // decrement remaning file capacity
			added = append(added, f) // save to added files list
			// don't bother checking the rest
			if b.Cap == 0 {
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

	if len(added) == len(files) {
		// if all files were successfully added
		log.Printf("[DEBUG] all files added to batch. remaining batch capacity (in bytes): %d", b.Cap)
		return added
	}
	// if we reach capacity before we finish with files,
	// return a list of the remaining files
	if b.Cap == MAX && len(added) < len(files) {
		return Diff(added, files)
	}
	// if b.Cap < MAX and we have left over files that were passed over for
	// being to large for the current batch
	if len(notAdded) > 0 && b.Cap < MAX {
		return notAdded
	}
	return nil
}
