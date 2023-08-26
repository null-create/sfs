package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"
)

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
