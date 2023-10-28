package service

import (
	"fmt"
	"log"
)

// max batch size
//
// (1,000,000,000 / 2^30 bytes/ 1 Gb)
const MAX int64 = 1e+9

/*
used to communicate the result of b.AddFiles() to the caller

	0 = failure
	1 = successful
	2 = left over files && b.Cap < 0
	3 = left over files && b.Cap == MAX
*/
type BatchStatus int

// status enums
const (
	Failure  BatchStatus = 0
	Success  BatchStatus = 1
	UnderCap BatchStatus = 2
	CapMaxed BatchStatus = 3
)

// used for keeping track of file additions and ommissions
// during b.AddFiles()
type AddCtx struct {
	Added    []*File
	NotAdded []*File
	Ignored  []*File
}

func NewCtx() *AddCtx {
	return &AddCtx{
		Added:    make([]*File, 0),
		NotAdded: make([]*File, 0),
		Ignored:  make([]*File, 0),
	}
}

// batch represents a collection of files to be uploaded or downloaded
type Batch struct {
	ID    string // batch ID (UUID)
	Cap   int64  // remaining capacity (in bytes)
	Max   int64  // total capacity (in bytes)
	Total int    // total files in this batch

	Files map[string]*File // files to be uploaded or downloaded
}

// create a new batch with capacity of MAX
func NewBatch() *Batch {
	return &Batch{
		ID:    NewUUID(),
		Cap:   MAX,
		Max:   MAX,
		Files: make(map[string]*File, 0),
	}
}

// used to prevent duplicate files from appearing in a batch
func (b *Batch) HasFile(id string) bool {
	if _, exists := b.Files[id]; exists {
		return true
	}
	return false
}

// retrieves all files in the batch.
//
// should be used when multiplexing file uploads or downloads
func (b *Batch) GetFiles() ([]*File, error) {
	if len(b.Files) == 0 {
		return nil, fmt.Errorf("no files were in the batch")
	}
	f := make([]*File, 0, len(b.Files))
	for _, file := range b.Files {
		f = append(f, file)
	}
	return f, nil
}

/*
we're bound  by an upper size limit on our batch sizes (MAX) since
we ideally don't want to clog a home network's resources when uploading
or downloading batches of files. MAX is subject to change of course,
but its in place as a mechanism for resource management.
*/
func (b *Batch) AddFiles(files []*File) ([]*File, BatchStatus) {
	// remember which ones we added so we don't have to modify the
	// files slice in-place as we're iterating over it
	//
	// remember which files were/weren't added or were ignored
	c := NewCtx()

	// take our unsorted slice of file objets and sort by size, then
	// return the results as Pairlist whos values are sorted by value in
	// ascending order
	sortedFiles := b.SortMapByValue(b.SliceToMap(files))

	for _, f := range sortedFiles {
		if !b.HasFile(f.File.ID) {
			// "if a current file's size doesn't exceed the remaining batch capacity, add it."
			//
			// this is basically a greedy approach, but that may change.
			//
			// EDIT: sorting has been added, but leaving the comments below because reasons
			//
			// since lists are unsorted, a file that is much larger than its neighbors may cause
			// batches to not contain as many possible files since one files weight may greatly tip
			// the scales, as it were. NP problems are hard.
			//
			// pre-sorting the list of files will introduce a lower O(nlogn) bound on any possible
			// resulting solution, so our current approach, which assumes the slice of
			// files is unsorted, falls around roughly O(nk) (i think), where n is the
			// number of times we need to iterate over the list of files (and remaning subsets after
			// each batch) and where k is the size of the *current* list we're iterating over and
			// building a batch from (assuming slice shrinkage with each pass).
			if b.Cap-f.File.Size() >= 0 {
				b.Files[f.File.ID] = f.File
				b.Cap -= f.File.Size()
				b.Total += 1
				c.Added = append(c.Added, f.File)
				if b.Cap == 0 { // don't bother checking the rest
					break
				}
			} else {
				// we want to check the other files in this list
				// since they may be small enough to add onto this batch.
				log.Printf("[DEBUG] file size (%d bytes) exceeds remaining batch capacity (%d bytes).\nattempting to add others...\n", f.File.Size(), b.Cap)
				c.NotAdded = append(c.NotAdded, f.File)
				continue
			}
		} else {
			log.Printf("[DEBUG] file (id=%s) already present. skipping...", f.File.ID)
			c.Ignored = append(c.Ignored, f.File)
			continue
		}
	}

	if len(c.Ignored) > 0 {
		log.Print("[DEBUG] there were duplicates in supplied file list.")
	}

	// success
	if len(c.Added) == len(files) {
		log.Printf("[DEBUG] all files added to batch. remaining batch capacity (in bytes): %d", b.Cap)
		return c.Added, Success
	}
	// if we reach capacity before we finish with files,
	// return a list of the remaining files
	if b.Cap == 0 && len(c.Added) < len(files) {
		log.Printf("[DEBUG] reached capacity before we could finish with the remaining files. \nreturning remaining files\n")
		return Diff(c.Added, files), CapMaxed
	}
	// if b.Cap < MAX and we have left over files that were passed over for
	// being to large for the current batch.
	if len(c.NotAdded) > 0 && b.Cap < b.Max {
		log.Printf("[DEBUG] returning files passed over for being too large for this batch")
		if len(c.Added) == 0 {
			log.Printf("[WARNING] *no* files were added!")
		}
		return c.NotAdded, UnderCap
	}

	return nil, 0
}

// used for adding single large files to a custom batch (doesn't care about MAX)
func (b *Batch) AddLgFiles(files []*File) (BatchStatus, error) {
	if len(files) == 0 {
		return Failure, fmt.Errorf("no files were added")
	}
	for _, f := range files {
		if !b.HasFile(f.ID) {
			b.Files[f.ID] = f
			b.Total += 1
		}
	}
	return Success, nil
}
