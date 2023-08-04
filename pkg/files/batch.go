package files

import "log"

// batch represents a collection of files to be uploaded or downloaded
// Batch.limit is set by the network profiler
type Batch struct {
	ID    string `json:"id"`    // batch ID (UUID)
	Limit int64  `json:"limit"` // total FILE SIZE in kb allowable for this batch.
	Cap   int64  `json:"cap"`   // remaining capacity (in kb)

	// files to be uploaded or downloaded
	Files []*File
}

func NewBatch(id string, limit int64) *Batch {
	return &Batch{
		ID:    id,
		Limit: limit,
	}
}

// NOTE this does not check whether there are any duplicate files in the
// files []*File slice... probably should do that
func (b *Batch) AddFiles(files []*File) []*File {
	added := make([]*File, len(files))
	for _, f := range files {
		if b.Cap-f.Size() <= 0 {
			log.Print("[DEBUG] batch limit reached with this file. attempting to add others...")
			// we want to check the other files in this list
			// since they may be small enough to add onto this batch.
			continue
		} else {
			b.Files = append(b.Files, f)
			b.Cap -= f.Size() // update remaning file capacity
			added = append(added, f)
		}
	}

	// TODO: if we reach our limit before we finish with files []*Files, we should
	// return a new list with the remaining files that weren't added to this batch
	if len(added) == len(files) {
		log.Printf("[DEBUG] all files added to batch. remaining capacity: %d", b.Limit-int64(len(b.Files)))
		return nil
	} else {
		log.Print("[DEBUG] batch wasn't able to add all files from supplied list. returning list of remaining files")
		// TODO: create ramaining files list
		// (compare added with files and store any that ARENT in added in a remaining files list)
	}
	return nil
}
