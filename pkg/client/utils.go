package client

import (
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/sfs/pkg/env"
)

func GetWd() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return wd
}

// Generate a pseudo-random integer in the range [0, n)
func RandInt(limit int) int {
	return rand.Intn(limit)
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

// converts a struct to a map.
// mainly so we can make a struct iterable.
func structToMap(s interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	v := reflect.ValueOf(s)
	t := reflect.TypeOf(s)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i).Interface()
		m[field.Name] = value
	}

	return m
}
