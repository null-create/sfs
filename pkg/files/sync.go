package files

import (
	"log"
	"time"
)

type SyncIndex struct {

	// We will use the file path for each file to retrieve the pointer for the
	// file object if it is to be queued for uploading or downloading
	// key = file path, value = last modified date
	LastSync map[string]time.Time

	// map of files to be queued for uploading or downloading
	// key = file UUID, value = file pointer
	ToUpdate map[string]*File
}

func NewSyncIndex() *SyncIndex {
	return &SyncIndex{
		LastSync: make(map[string]time.Time),
		ToUpdate: make(map[string]*File),
	}
}

/*
get a slice of file paths from the SyncIndex.ToUpdate map

can be used when generating lists of files to be processed for uploading or downloading
*/
func (s *SyncIndex) GetFilePaths() []string {
	if len(s.ToUpdate) == 0 {
		log.Printf("[DEBUG] no files queued for uploading or downloading")
		return nil
	}

	fp := make([]string, 0)
	for _, file := range s.ToUpdate {
		fp = append(fp, file.Path)
	}
	return fp
}

/*
build a new sync index starting with a given directory which
is treated as the "root" of our inquiry. all subdirectories will be checked,
but we assume this is the root, and that there is no parent directory!

utilizes the directory's d.WalkS() function
*/
func BuildSyncIndex(dir *Directory) *SyncIndex {
	idx := dir.WalkS()
	if idx == nil {
		log.Fatalf("[ERROR] could not build sync index for dir: %v", dir.ID)
	}
	return idx
}

/*
takes a given directory pointer and compares it against against a sync index's
internal LastSync map. it's assumed the index was created before this function was called.

if the sync time in the last sync map is less recent than whats in the current directory, then we add that file to the ToUpdate map,
which will be used to create a file upload or download queue
*/
func BuildToUpdate(dir *Directory, idx *SyncIndex) *SyncIndex {
	if idx == nil {
		log.Fatalf("[ERROR] could not build ToUpdate index; idx was nil")
	}
	for _, f := range dir.Files {
		// check if the time difference between most recent sync and last sync
		// is greater than zero.
		if f.LastSync.Sub(idx.LastSync[f.Path]) > 0 {
			idx.ToUpdate[f.ID] = f
		}
	}
	if len(idx.ToUpdate) == 0 {
		log.Printf("[DEBUG] no files matched for syncing")
		return nil
	}
	return idx
}
