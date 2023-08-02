package files

import "log"

// batch represents a collection of files to be uploaded or downloaded
// Batch.limit is set by the network profiler
type Batch struct {
	ID    string `json:"id"`    // batch ID (UUID)
	Limit int    `json:"limit"` // total number of files to be uploaded or downloaded
	Cap   int    `json:"cap"`   // remaining capacity

	// files to be uploaded or downloaded
	Files []*File
}

func NewBatch(id string, limit int) *Batch {
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
		// check if this latest file will cause us to reach our limit
		if len(b.Files)+1 > b.Limit {
			log.Print("[DEBUG] batch limit reached. exiting...")
			break
		} else {
			b.Files = append(b.Files, f)
			added = append(added, f)
		}
	}

	// TODO: if we reach our limit before we finish with files []*Files, we should
	// return a new list with the remaining files that weren't added to this batch
	if len(added) == len(files) {
		log.Printf("[DEBUG] all files added to batch. remaining capacity: %d", b.Limit-len(b.Files))
		return nil
	} else {
		log.Print("[DEBUG] batch wasn't able to add all files from supplied list. returning list of remaining files")
		// TODO: create ramaining files list
	}
	return nil
}
