package service

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type SyncIndex struct {
	// userID of of the user this sync index belongs to
	UserID string `json:"user"`

	// filepath to save sync-index.json to, i.e.
	// path/to/userID-sync-index-date.json, where <date>
	// is updated with each save
	IdxFp string `json:"file"`

	// flag to indicate whether a sync operation should be executed
	Sync bool `json:"sync"`

	// We will use the file path for each file to retrieve the pointer for the
	// file object if it is to be queued for uploading or downloading
	//
	// key = file UUID, value = last modified date
	LastSync map[string]time.Time `json:"last_sync"`

	// map of files to be queued for uploading or downloading
	//
	// key = file UUID, value = file pointer
	ToUpdate map[string]*File `json:"to_update"`
}

// create a new sync-index object
func NewSyncIndex(userID string) *SyncIndex {
	return &SyncIndex{
		UserID:   userID,
		IdxFp:    "",
		Sync:     false,
		LastSync: make(map[string]time.Time, 0),
		ToUpdate: make(map[string]*File, 0),
	}
}

func (s *SyncIndex) isInit() bool {
	if s.ToUpdate == nil {
		return false
	}
	if s.LastSync == nil {
		return false
	}
	return true
}

func (s *SyncIndex) IsMapped() bool {
	if !s.isInit() {
		return false
	}
	if len(s.LastSync) == 0 {
		return false
	}
	if len(s.ToUpdate) == 0 {
		return false
	}
	return true
}

// write out a sync index to a JSON file
func (s *SyncIndex) SaveToJSON() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	fn := fmt.Sprintf("%s-sync-index-%s.json", s.UserID, time.Now().Format("2006-01-02T15-04-05"))
	return os.WriteFile(filepath.Join(s.IdxFp, fn), data, 0644)
}

// checks last sync for file.
// won't be in toupdate if it's not in lastsync first
func (s *SyncIndex) HasFile(fileID string) bool {
	if _, exists := s.LastSync[fileID]; !exists {
		return false
	}
	return true
}

// get a slice of files to sync from the index.ToUpdate map
func (s *SyncIndex) GetFiles() []*File {
	if len(s.ToUpdate) == 0 {
		log.Print("no files matched for syncing")
		return nil
	}
	syncFiles := make([]*File, 0, len(s.ToUpdate))
	for _, f := range s.ToUpdate {
		syncFiles = append(syncFiles, f)
	}
	return syncFiles
}

// ----------------------------------------------------------------

/*
build a new sync index starting with a given directory which
is treated as the "root" of our inquiry. all subdirectories will be checked,
but we assume this is the root, and that there is no parent directory!

utilizes the directory's d.WalkS() function
*/
func BuildSyncIndex(root *Directory) *SyncIndex {
	idx := NewSyncIndex(root.Owner)
	if idx := root.WalkS(idx); idx != nil {
		return idx
	}
	return nil
}

/*
takes a given directory pointer and compares it against against a sync index's
internal LastSync map. it's assumed the index was created before this function was called.

if the sync time in the last sync map is less recent than whats in the current directory, then we add that file to the ToUpdate map,
which will be used to create a file upload or download queue
*/
func BuildToUpdate(root *Directory, idx *SyncIndex) *SyncIndex {
	if idx := root.WalkU(idx); idx != nil {
		return idx
	}
	return nil
}

// if all files in the given slice are greater than
// the current capacity of this batch, then none of them
// will be able to be added
func wontFit(files []*File, limit int64) bool {
	var total int
	for _, f := range files {
		if f.Size() > limit {
			total += 1
		}
	}
	return total == len(files)
}

// generates a slice of files that are all under MAX,
// from a raw list of files
func Prune(files []*File) []*File {
	lgf := GetLargeFiles(files)
	return Diff(files, lgf)
}

// build a slice of file objects that exceed batch.MAX.
//
// these can be added to a custom batch to be uploaded/downloaded
// after the ordinary batch queue is done processing
func GetLargeFiles(files []*File) []*File {
	if len(files) == 0 {
		return []*File{}
	}
	f := make([]*File, 0, len(files))
	for _, file := range files {
		if file.Size() > MAX {
			f = append(f, file)
		}
	}
	return f
}

// keep adding the left over files to new batches until
// have none left over from each b.AddFiles() call
func buildQ(f []*File, b *Batch, q *Queue) *Queue {
	for len(f) > 0 {
		lo, status := b.AddFiles(f)
		switch status {
		case Success:
			q.Enqueue(b)
			return q
		case CapMaxed:
			q.Enqueue(b)
			nb := NewBatch()
			b = nb
		case UnderCap:
			// if none of the left over files will fit in the
			// current batch, create a new one and move on
			if wontFit(lo, b.Cap) {
				if len(b.Files) > 0 {
					q.Enqueue(b)
					nb := NewBatch()
					b = nb
				}
			}
		}
		f = lo
	}
	return q
}

// build the queue for file uploads or downloads during a Sync event
//
// idx should have ToUpdate populated
func BuildQ(idx *SyncIndex) *Queue {
	files := idx.GetFiles()
	if files == nil {
		return nil
	}
	if len(files) == 0 {
		log.Printf("[DEBUG] no files matched for syncing")
		return nil
	}
	// if every individual file exceeds b.MAX, none will able to
	// be added to the standard batch queue and we like to avoid infinite loops,
	// so we need to create a large file queue
	if wontFit(files, MAX) {
		log.Print("[WARNING] all files exceeded b.MAX. creating large file queue")
		return LargeFileQ(files)
	}
	return buildQ(files, NewBatch(), NewQ())
}

// create a custom file queue for files that exceed batch.MAX
func LargeFileQ(files []*File) *Queue {
	b := NewBatch()
	b.AddLgFiles(files)
	q := NewQ()
	q.Enqueue(b)
	return q
}
