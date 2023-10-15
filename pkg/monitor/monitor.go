package monitor

/*
this is the file for the background event-listener daemon.

this will listen for events like a file being saved within the client's drive, which will then
automatically start a new sync index operation. whether the user wants to automatically sync or not
should be a setting, but the daemon will automatically make a new sync index with each file or directory
modification.

should also have a mechanism to interrupt a sync operation if a new event occurs.
*/

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// see: https://medium.com/@skdomino/watch-this-file-watching-in-go-5b5a247cf71f

type Monitor struct {
	// path to the directory to monitor
	Path string

	// watcher for file and directory events
	Watcher *fsnotify.Watcher
}

// NOTE: must call watcher close after instantiation@
func NewMonitor(path string) *Monitor {
	watcher, err := fsnotify.NewWatcher() // Create new watcher.
	if err != nil {
		log.Fatal(err)
	}
	return &Monitor{
		Path:    path,
		Watcher: watcher,
	}
}

func (m *Monitor) watchDir(path string, fi os.FileInfo, err error) error {
	// since fsnotify can watch all the files in a directory, watchers only need
	// to be added to each nested directory
	if fi.Mode().IsDir() {
		return m.Watcher.Add(path)
	}
	return nil
}

func (m *Monitor) MonitorDrive(drvPath string) {
	defer m.Watcher.Close()

	// starting at the root of the drive, walk each file/directory searching for
	// directories
	if err := filepath.Walk(drvPath, m.watchDir); err != nil {
		fmt.Println("ERROR", err)
	}

	// shutdown channel
	done := make(chan bool)

	// start listening for events
	go func() {
		for {
			select {
			case event, ok := <-m.Watcher.Events:
				if !ok {
					log.Printf("[WARNING] monitoring failed: %v", event)
					return
				}
				// TODO: parse various events to signal sync events or
				// other index building operations
				log.Println("[INFO] event:", event)
				if event.Has(fsnotify.Write) {
					log.Println("modified file:", event.Name)
				} else if event.Has(fsnotify.Create) {

				} else if event.Has(fsnotify.Remove) {

				} else if event.Has(fsnotify.Write) {

				}
			case err, ok := <-m.Watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	// add a path
	err := m.Watcher.Add(drvPath)
	if err != nil {
		log.Fatal(err)
	}

	<-done
}
