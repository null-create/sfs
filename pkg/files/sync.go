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
	// user this sync index belongs to
	User   string `json:"user"`
	UserID string `json:"user_id"`

	// filepath to save sync-index.json to, i.e.
	// path/to/userID-sync-index-date.json, where <date>
	// is updated with each save
	IdxFp string `json:"file"`

	// We will use the file path for each file to retrieve the pointer for the
	// file object if it is to be queued for uploading or downloading
	// key = file UUID, value = last modified date
	LastSync map[string]time.Time `json:"last_sync"`

	// map of files to be queued for uploading or downloading
	// key = file UUID, value = file pointer
	ToUpdate map[string]*File `json:"to_update"`
}

// create a new sync-index object
func NewSyncIndex(user, userID string) *SyncIndex {
	return &SyncIndex{
		User:     user,
		UserID:   userID,
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
	fn := fmt.Sprintf("%s-sync-index-%s.json", s.UserID, time.Now().Format("00-00-00-01-02-2006"))
	return os.WriteFile(filepath.Join(s.IdxFp, fn), data, 0644)
}

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

// get a slice of files to sync
func (s *SyncIndex) GetFiles() ([]*File, error) {
	if len(s.ToUpdate) == 0 {
		return nil, fmt.Errorf("no files matched for syncing")
	}
	syncFiles := make([]*File, 0, len(s.ToUpdate))
	for _, f := range s.ToUpdate {
		syncFiles = append(syncFiles, f)
	}
	return syncFiles, nil
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

// if all files are above MAX, then none of these files
// be able to be added to a batch
func willOverflow(files []*File) bool {
	if len(files) == 0 {
		return false
	}
	var total int
	for _, f := range files {
		if f.Size() > MAX {
			total += 1
		}
	}
	// if all files are above MAX,
	// then none will be added to a batch
	return total == len(files)
}

// TODO: remove any files from the queue that individually exceed MAX?
// create a separate queue for single large files?

// prepare a slice of batches to be queued for uploading or downloading
//
// populates from idx.ToUpdate
func Sync(root *Directory, idx *SyncIndex) (*Queue, error) {
	if len(idx.ToUpdate) == 0 {
		return nil, fmt.Errorf("no files found to sync for root %s", root.ID)
	}
	files, err := idx.GetFiles()
	if err != nil {
		return nil, err
	}
	// if every individual file exceeds b.MAX, none will able to be added,
	// so we want to avoid infinite recursive calls/stack overflows
	if willOverflow(files) {
		return nil, fmt.Errorf("all files exceeded max batch size limit. none can be added to queue")
	}
	queue := NewQ()
	return buildFileQ(files, queue)
}

// recursively builds a batch queue from a list of files.
//
// the queue will be used to upload/download files to the server
func buildFileQ(files []*File, q *Queue) (*Queue, error) {
	if len(files) == 0 {
		return q, nil
	}
	b := NewBatch()
	f := b.AddFiles(files)
	q.Enqueue(b)
	// there were left over files that
	// didn't make it into this batch
	if len(f) > 0 && len(f) < len(files) {
		return buildFileQ(f, q)
		// all files were added to the batch
	} else if len(f) == len(files) {
		return q, nil
	}
	return q, nil
}
