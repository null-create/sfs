package files

import (
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
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

// creates a new UUID string
func NewUUID() string {
	uuid, err := uuid.NewUUID()
	if err != nil {
		log.Fatalf("[ERROR] failed to generate UUID: \n%v\n", err)
	}
	return uuid.String()
}

// return the difference between two []*File slices.
//
// assuming that go's map implementation has ~O(1) access time,
// then this function should work in ~O(n) on an unsorted slice.
//
// https://stackoverflow.com/questions/19374219/how-to-find-the-difference-between-two-slices-of-strings
func Diff(f, g []*File) []*File {
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
