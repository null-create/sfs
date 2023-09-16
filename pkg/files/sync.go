package files

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
	User string `json:"user"`

	// filepath to save sync-index.json to, i.e.
	// path/to/userID-sync-index-date.json, where <date>
	// is updated with each save
	IdxFp string `json:"file"`

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
func NewSyncIndex(user string) *SyncIndex {
	return &SyncIndex{
		User:     user,
		LastSync: make(map[string]time.Time, 0),
		ToUpdate: make(map[string]*File, 0),
	}
}

// write out a sync index to a JSON file
func (s *SyncIndex) SaveToJSON() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	fn := fmt.Sprintf("%s-sync-index-%s.json", s.User, time.Now().Format("00-00-00-01-02-2006"))
	return os.WriteFile(filepath.Join(s.IdxFp, fn), data, 0644)
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

// get a slice of file paths from the SyncIndex.ToUpdate map
//
// can be used when generating lists of files to be processed for uploading or downloading
func (s *SyncIndex) GetFilePaths() []string {
	if len(s.ToUpdate) == 0 {
		log.Printf("[DEBUG] no files queued for uploading or downloading")
		return []string{}
	}
	fp := make([]string, 0, len(s.ToUpdate))
	for _, file := range s.ToUpdate {
		fp = append(fp, file.ServerPath)
	}
	return fp
}

// ----------------------------------------------------------------

/*
build a new sync index starting with a given directory which
is treated as the "root" of our inquiry. all subdirectories will be checked,
but we assume this is the root, and that there is no parent directory!

utilizes the directory's d.WalkS() function
*/
func BuildSyncIndex(root *Directory) *SyncIndex {
	if idx := root.WalkS(); idx != nil {
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
func prune(files []*File) []*File {
	lgf := getLargeFiles(files)
	return Diff(files, lgf)
}

// build a slice of file objects that exceed batch.MAX.
//
// these can be added to a custom batch to be uploaded/downloaded
// after the ordinary batch queue is done processing
func getLargeFiles(files []*File) []*File {
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

// create a custom file queue for files that exceed batch.MAX
func LargeFileQ(files []*File) *Queue {
	b := NewBatch()
	b.AddLgFiles(files)
	q := NewQ()
	q.Enqueue(b)
	return q
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
				q.Enqueue(b)
				nb := NewBatch()
				b = nb
			}
		}
		f = lo
	}
	return q
}

// build the queue for file uploads or downloads during a Sync event
func BuildQ(idx *SyncIndex, q *Queue) (*Queue, error) {
	files := idx.GetFiles()
	if files == nil {
		return nil, fmt.Errorf("no files found to sync for syncing")
	}
	// if every individual file exceeds b.MAX, none will able to
	// be added to the standard batch queue, and we like to avoid infinite loops,
	// so we need to create a custom large file queue
	if wontFit(files, MAX) {
		log.Print("[WARNING] all files exceeded b.MAX. creating custom large file queue")
		return LargeFileQ(files), nil
	}
	return buildQ(files, NewBatch(), NewQ()), nil
}
