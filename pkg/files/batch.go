package files

import "log"

// max batch size (1gb / 1,073,741,824 bytes)
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

// NOTE this does not check whether there are any duplicate files in the
// files []*File slice... probably should do that
func (b *Batch) AddFiles(files []*File) []*File {
	added := make([]*File, len(files))
	for _, f := range files {
		if b.Cap-f.Size() <= 0 {
			// we want to check the other files in this list
			// since they may be small enough to add onto this batch.
			log.Printf("[DEBUG] file size (%d bytes) exceeds remaining batch capacity (%d bytes).\nattempting to add others...\n", f.Size(), b.Cap)
			continue
		} else {
			b.Files = append(b.Files, f)
			b.Cap -= f.Size()        // update remaning file capacity
			added = append(added, f) // save to added files list
		}
	}

	// TODO: if we reach our limit before we finish with files []*Files, we should
	// return a new list with the remaining files that weren't added to this batch
	if len(added) == len(files) {
		log.Printf("[DEBUG] all files added to batch. remaining batch capacity (in bytes): %d", b.Cap)
		return nil
	} else if len(added) < len(files) {
		log.Print("[DEBUG] batch wasn't able to add all files from supplied list. returning list of remaining files")
		// TODO: capture remaining files
		// (files-added) = remaining files list to be returned
	}
	return added
}
