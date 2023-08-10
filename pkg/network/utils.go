package network

import (
	"encoding/json"
	"log"
	"os"
)

const PROFILE string = "network-profile.json"

func GetHostName() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("[ERROR] unable to get hostname \n%v\n ", err)
		return ""
	}
	return hostname
}

// SaveProfile creates a simple .json file of our network speed averages
func SaveProfile(profile *NetworkProfile) error {

	// TODO: create dedicated service file path for network profile

	// Open the JSON file for writing
	file, err := os.Create(PROFILE)
	if err != nil {
		log.Fatalf("[ERROR] error creating file \n%v\n ", err)
	}
	defer file.Close()

	// Encode the data and write it to the file
	encoder := json.NewEncoder(file)
	jsonStr, err := json.Marshal(&profile)
	if err != nil {
		log.Fatalf("[ERROR] error marshalling data to JSON format \n%v\n ", err)
	}
	err = encoder.Encode(jsonStr)
	if err != nil {
		log.Fatalf("[ERROR] error encoding data to JSON \n%v\n ", err)
	}

	return nil
}
