package files

import (
	"log"
	"time"
)

type SyncIndex struct {

	// We will use the file path for each file to retrieve the pointer for the
	// file object if it is to be queued for uploading or downloading
	// key = file UUID, value = last modified date
	LastSync map[string]time.Time

	// map of files to be queued for uploading or downloading
	// key = file UUID, value = file pointer
	ToUpdate map[string]*File
}

func NewSyncIndex() *SyncIndex {
	return &SyncIndex{
		LastSync: make(map[string]time.Time, 0),
		ToUpdate: make(map[string]*File, 0),
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
		fp = append(fp, file.ServerPath)
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
	if idx := dir.WalkS(); idx != nil {
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
func BuildToUpdate(dir *Directory, idx *SyncIndex) *SyncIndex {
	if idx := dir.WalkU(idx); idx != nil {
		return idx
	}
	return nil
}
