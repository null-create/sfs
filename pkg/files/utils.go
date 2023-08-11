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
		log.Fatalf("[ERROR] Could not generate UUID:\n%v\n", err)
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
	tmp := make(map[*File]string, len(g))
	for _, file := range g {
		tmp[file] = "hi"
	}
	var diff []*File
	for _, file := range f {
		if _, found := tmp[file]; !found {
			diff = append(diff, file)
		}
	}
	return diff
}
