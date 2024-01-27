package service

import (
	"fmt"
	"log"
	"path/filepath"
)

// --------- sync between hard drives ----------------

// get paths for source and destination disks
// create a sync index of the source disk, then
// build a queue for physical *copying*, not tranferring
// via HTTP.
func SyncDisks(srcDir *Directory, destRootDir *Directory) error {
	// build sync index
	srcIdx := BuildSyncIndex(srcDir)
	if srcIdx == nil {
		return fmt.Errorf("failed to create sync index from source disk")
	}
	srcIdx = BuildToUpdate(srcDir, srcIdx)
	// build sync queue
	queue := BuildQ(srcIdx)
	if queue == nil || len(queue.Queue) == 0 {
		log.Print("[INFO] no files matched for syncing. exiting sync.")
		return nil
	}
	// copy files
	for _, batch := range queue.Queue {
		for _, file := range batch.Files {
			go func() {
				// TODO: maybe add separate disk field in file struct?
				// NOTE: each file may have a different directory for their destination
				// on the other disk. may have to find dest sub directory and treat the
				// argument destDir as rootDir for the destinaton disc.
				destPath := filepath.Join(destRootDir.Path, file.Name)
				if err := file.Copy(destPath); err != nil {
					log.Printf("[WARNING] failed to copy file (id=%s) \n%v", file.ID, err)
				}
				// update dir struct
				if err := destRootDir.AddFile(file); err != nil {
					log.Printf("[WARNING] failed to update dir: %v", err)
				}
			}()
		}
	}
	// TODO: verify that the files were copied correctly
	// srcFiles := srcIdx.GetFiles()
	return nil
}

// a local daemon that listens for file events and coordinates synchronization
// operations between two hard disks. this is to implement a specific SFS mode that
// doesn't depend on the client/server configuration, and instead uses the user's
// local file systems/hard disks
func DiskDaemon() error { return nil }
