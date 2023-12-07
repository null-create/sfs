package client

import (
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"

	svc "github.com/sfs/pkg/service"
)

func GetWd() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return wd
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

// parse the NEW_SERVICE env var to see if we are instantiating a new sfs service
func isMode(k string) bool {
	env := svc.NewE()
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
func Exists(filename string) bool {
	if _, err := os.Stat(filename); err != nil && err == os.ErrNotExist {
		return false
	} else if err != nil && err != os.ErrNotExist {
		log.Fatalf("unable to get file status: %v", err)
	}
	return true
}
