package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/db"
	"github.com/sfs/pkg/files"
)

// Generate a random integer in the range [1, n)
func RandInt(limit int) int {
	// Seed the random number generator with the current time
	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(limit)
	if num == 0 { // don't want zero so we just autocorrect to 1 if that happens
		return 1
	}
	return num
}

func isEmpty(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	if len(entries) == 0 {
		return true
	}
	return false
}

// write out as a json file
func saveJSON(dir, filename string, data map[string]interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalf("[ERROR] failed marshalling JSON data: %s\n", err)
	}

	if err = ioutil.WriteFile(filepath.Join(dir, filename), jsonData, 0666); err != nil {
		log.Fatalf("[ERROR] unable to write JSON file %s: %s\n", filename, err)
	}
}

// ----- db utils --------------------------------

// get file info from db
func findFile(fileID string, dbDir string) (*files.File, error) {
	q := db.NewQuery(dbDir, false)
	f, err := q.GetFile(fileID)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func findUser(userID string, dbDir string) (*auth.User, error) {
	q := db.NewQuery(dbDir, false)
	u, err := q.GetUser(userID)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func findDir(dirID string, dbDir string) (*files.Directory, error) {
	q := db.NewQuery(dbDir, false)
	d, err := q.GetDirectory(dirID)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func findDrive(driveID string, dbDir string) (*files.Drive, error) {
	q := db.NewQuery(dbDir, false)
	d, err := q.GetDrive(driveID)
	if err != nil {
		return nil, err
	}
	return d, nil
}
