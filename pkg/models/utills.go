package models

import (
	"encoding/json"
	"log"
)

// write out as a json file
func ToJSON(obj any) []byte {
	json, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		log.Fatalf("[ERROR] failed marshalling JSON data: %s\n", err)
	}
	return json
}
