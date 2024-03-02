package server

import (
	"log"
	"math/rand"
	"os"
	"time"
)

// Generate a random integer in the range [1, n)
func RandInt(n int) int {
	// Seed the random number generator with the current time
	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(n)
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

func secondsToTimeStr(seconds float64) string {
	duration := time.Duration(int64(seconds)) * time.Second
	timeValue := time.Time{}.Add(duration)
	return timeValue.Format("15:04:05")
}
