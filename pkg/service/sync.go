package service

import (
	"encoding/json"
	"log"
	"time"
)

/*
A SyncIndex is a data structure usued coordinate which files should be updated,
removed, or added between the client and server (or between hard disks,
if this is the service mode that's active)
*/
type SyncIndex struct {
	// userID of of the user this sync index belongs to
	UserID string `json:"user"`

	// flag to indicate whether a sync operation should be executed
	Sync bool `json:"sync"`

	// We will use the file path for each file to retrieve the pointer for the
	// file object if it is to be queued for uploading or downloading
	//
	// key = file or directory UUID, value = last modified date
	LastSync map[string]time.Time `json:"last_sync"`

	// map of files to be queued for uploading or downloading.
	// key = file UUID, value = file pointer
	FilesToUpdate map[string]*File `json:"files_to_update"`

	// NOTE: no longer used since directories aren't being monitored.
	// map of directories to be queued for uploading or downloading
	// key = dir UUID, value = dir pointer
	// DirsToUpdate map[string]*Directory `json:"dirs_to_update"`
}

// create a new sync-index object
func NewSyncIndex(userID string) *SyncIndex {
	return &SyncIndex{
		UserID:        userID,
		Sync:          false,
		LastSync:      make(map[string]time.Time, 0),
		FilesToUpdate: make(map[string]*File, 0),
		// DirsToUpdate:  make(map[string]*Directory, 0),
	}
}

// resets both LastSync and ToUpdate maps
func (s *SyncIndex) Reset() {
	s.LastSync = nil
	s.FilesToUpdate = nil
	// s.DirsToUpdate = nil
	s.LastSync = make(map[string]time.Time, 0)
	s.FilesToUpdate = make(map[string]*File, 0)
	// s.DirsToUpdate = make(map[string]*Directory, 0)
}

// converts to json format for transfer
func (s *SyncIndex) ToJSON() ([]byte, error) {
	data, err := json.MarshalIndent(s, "", " ")
	if err != nil {
		return nil, err
	}
	return data, nil
}

// checks last sync for file or directory.
// won't be in toupdate if it's not in lastsync first.
func (s *SyncIndex) HasItem(itemId string) bool {
	if _, exists := s.LastSync[itemId]; exists {
		return true
	}
	return false
}

// make a json-formatted string representation of the sync-index object
func (s *SyncIndex) ToString() string {
	data, err := s.ToJSON()
	if err != nil {
		log.Fatal(err)
	}
	return string(data)
}

// get a slice of files to sync from the index.ToUpdate map
func (s *SyncIndex) GetFiles() []*File {
	if len(s.FilesToUpdate) == 0 {
		log.Print("no files matched for syncing")
		return nil
	}
	syncFiles := make([]*File, 0, len(s.FilesToUpdate))
	for _, f := range s.FilesToUpdate {
		syncFiles = append(syncFiles, f)
	}
	return syncFiles
}

// ------ index building ----------------------------------------------

/*
build a new sync index starting with a given directory which
is treated as the "root" of our inquiry. all subdirectories will be checked,
but we assume this is the root, and that there is no parent directory!

utilizes the directory's d.WalkS() function
*/
func BuildRootSyncIndex(root *Directory) *SyncIndex {
	idx := NewSyncIndex(root.OwnerID)
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
func BuildRootToUpdate(root *Directory, idx *SyncIndex) *SyncIndex {
	if idx := root.WalkU(idx); idx != nil {
		return idx
	}
	return nil
}

/*
compares a given syncindex against a newly generated one and returns the differnece
between the two, favoring the newer one for any last sync times.

the map this returns will only contain the itemps that were matched and found to have a
more recent time -- items that weren't matched will be ignored.

TODO: indicate whether the latest sync time is from a server item or a client item
*/
func Compare(orig *SyncIndex, new *SyncIndex) *SyncIndex {
	// index containing most recent items
	newest := NewSyncIndex(orig.UserID)

	// compare last sync times
	for itemId, lastSync := range new.LastSync {
		if origTime, exists := orig.LastSync[itemId]; exists {
			if lastSync.After(origTime) {
				newest.LastSync[itemId] = lastSync
			}
		}
	}
	// compare files marked for updating
	for fileID, newFile := range new.FilesToUpdate {
		if origFile, exists := orig.FilesToUpdate[fileID]; exists {
			if newFile.LastSync.After(origFile.LastSync) {
				newest.FilesToUpdate[fileID] = newFile
			}
		}
	}
	// compare directories marked for updating
	// for dirID, newDir := range new.DirsToUpdate {
	// 	if origDir, exists := orig.DirsToUpdate[dirID]; exists {
	// 		if newDir.LastSync.After(origDir.LastSync) {
	// 			diff.DirsToUpdate[dirID] = newDir
	// 		}
	// 	}
	// }
	return newest
}

/*
Build a sync index of all files being monitored by the system.

NOTE: the directories argument is for future implementations.
probably won't be used during this first iteration.
dirs can be set to nil for the time being.
*/
func BuildSyncIndex(files []*File, dirs []*Directory, idx *SyncIndex) *SyncIndex {
	for _, file := range files {
		if !idx.HasItem(file.ID) {
			idx.LastSync[file.ID] = file.LastSync
		} else {
			if file.LastSync.After(idx.LastSync[file.ID]) {
				idx.LastSync[file.ID] = file.LastSync
			}
		}
	}
	// NOTE: for future implementation iterations
	// for _, dir := range dirs {
	// 	if !idx.HasItem(dir.ID) {
	// 		idx.LastSync[dir.ID] = dir.LastSync
	// 	} else {
	// 		if dir.LastSync.After(idx.LastSync[dir.ID]) {
	// 			idx.LastSync[dir.ID] = dir.LastSync
	// 		}
	// 	}
	// }
	return idx
}

/*
update the indexes FilesToUpdate map of monitored distrtributed files.
assumes the supplied index's LastSync map is instantiated and populated, otherwise
will fail or give inaccurate results.

if item is not known to the index, then it will be ignored.
*/
func BuildToUpdate(files []*File, dirs []*Directory, idx *SyncIndex) *SyncIndex {
	for _, file := range files {
		if idx.HasItem(file.ID) {
			if file.LastSync.After(idx.LastSync[file.ID]) {
				idx.FilesToUpdate[file.ID] = file
			}
		}
	}
	// NOTE: for future implementation iterations
	// for _, d := range dirs {
	// 	if idx.HasItem(d.ID) {
	// 		if d.LastSync.After(idx.LastSync[d.ID]) {
	// 			idx.DirsToUpdate[d.ID] = d
	// 		}
	// }
	return idx
}

// ------- transfers --------------------------------

// if all files in the given slice are greater than
// the current capacity of this batch, then none of them
// will be able to be added
func wontFit(files []*File, limit int64) bool {
	var total int
	for _, f := range files {
		if f.GetSize() > limit {
			total += 1
		}
	}
	return total == len(files)
}

// generates a slice of files that are all under MAX,
// from a raw list of files
func Prune(files []*File) []*File {
	lgf := GetLargeFiles(files)
	return DiffFiles(files, lgf)
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
		if file.GetSize() > MAX {
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
//
// returns nil if no files are found from the given index
func BuildQ(idx *SyncIndex) *Queue {
	files := idx.GetFiles()
	if files == nil {
		return nil
	}
	// if every individual file exceeds b.MAX, none will able to
	// be added to the standard batch queue and we like to avoid infinite loops,
	// so we'll need to just create a large file queue instead.
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
