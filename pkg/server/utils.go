package server

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/sfs/pkg/env"
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

// parse the NEW_SERVICE env var to see if we are instantiating a new sfs service
func isMode(k string) bool {
	env := env.NewE()
	v, err := env.Get(k)
	if err != nil {
		log.Fatalf("failed to get %s env var: %v", k, err)
	}
	isMode, err := strconv.ParseBool(v)
	if err != nil {
		log.Fatalf("failed to parse env var string to bool: %v", err)
	}
	return isMode
}

// write out as a json file
func saveJSON(dir, filename string, data any) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalf("[ERROR] failed marshalling JSON data: %s\n", err)
	}

	if err = os.WriteFile(filepath.Join(dir, filename), jsonData, 0666); err != nil {
		log.Fatalf("[ERROR] unable to write JSON file %s: %s\n", filename, err)
	}
}
