package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

func KbToMb(kb float64) float64 {
	return kb / 1024.0
}

// Generate a random integer in the range [1, n)
func RandInt(limit int) int {
	// Seed the random number generator with the current time
	rand.Seed(time.Now().UnixNano())

	num := rand.Intn(limit)
	if num == 0 {
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

	if err = os.WriteFile(filepath.Join(dir, filename), jsonData, 0666); err != nil {
		log.Fatalf("[ERROR] unable to write JSON file %s: %s\n", filename, err)
	}
}

func GetCwd() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("[ERROR] unable to get current working directory %v", err)
	}
	return dir
}

// return the difference between two []*File slices.
//
// assuming that go's map implementation has ~O(1) access time,
// then this function should work in ~O(n) on an unsorted slice.
//
// based off of: https://stackoverflow.com/questions/19374219/how-to-find-the-difference-between-two-slices-of-strings
func DiffFiles(f, g []*File) []*File {
	tmp := make(map[*File]int)
	for _, file := range g {
		tmp[file] = 1
	}
	var diff []*File
	for _, file := range f {
		if _, found := tmp[file]; !found {
			diff = append(diff, file)
		}
	}
	return diff
}

// return the difference between two map[string]*Directory maps.
// compares the directory ID's.
func DiffDirs(f, g map[string]*Directory) []*Directory {
	tmp := make(map[string]bool)
	for _, dir := range g {
		tmp[dir.ID] = true
	}
	var diff []*Directory
	for _, dir := range f {
		if _, found := tmp[dir.ID]; !found {
			diff = append(diff, dir)
		}
	}
	return diff
}

// remove duplicate file pointers from a slice
//
// based off of: https://stackoverflow.com/questions/66643946/how-to-remove-duplicates-strings-or-int-from-slice-in-go
func RemDup(f []*File) []*File {
	tmp := make(map[*File]bool)
	res := make([]*File, 0)
	for _, file := range f {
		if _, found := tmp[file]; !found {
			tmp[file] = true
			res = append(res, file)
		}
	}
	return res
}

/*
get keys from the map -- faster than using append()
see: https://stackoverflow.com/questions/21362950/getting-a-slice-of-keys-from-a-map
func GetKeys(mymap map[T]T) []T {
	keys := make([]T, len(mymap))
	i := 0
	for k := range mymap {
    keys[i] = k
    i++
	}
	return keys
}
*/

// copy a file
func Copy(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	if err != nil {
		return err
	}
	return nil
}

// check if a file exists
func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

// create a directory if it doesn't exist yet.
func CreateIfNotExists(dir string, perm os.FileMode) error {
	if Exists(dir) {
		return nil
	}
	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}
	return nil
}

// copy a symbolic link
func CopySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return os.Symlink(link, dest)
}
