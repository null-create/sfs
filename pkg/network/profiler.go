package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// max number of iterations for measureSpeed()
const MAX = 100

// We measure network resources by timing how long it takes to download a file of
// size N to URL, then how long it takes to upload data of the same size to said URL.
// This function is used iteratively to measure the average upload and download times,
// which helps the server determine file batch sizes
type NetworkProfile struct {
	HostName string `json:"host"`

	BatchMAX int64 `json:"batch_max"`

	UpRate   float64 `json:"up_rate"`
	DownRate float64 `json:"down_rate"`
}

func NewNetworkProfile() *NetworkProfile {
	return &NetworkProfile{
		HostName: GetHostName(),
	}
}

func measureSpeed(url string, client *http.Client) (downloadSpeed, uploadSpeed float64, err error) {
	//------Measure download speed
	start := time.Now()

	resp, err := client.Get(url)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get URL: %v", err)
	}
	defer resp.Body.Close()

	downloadDuration := time.Since(start).Seconds()
	downloadSpeed = float64(resp.ContentLength) / downloadDuration

	//-----Measure upload speed with larger text file
	testFolder := filepath.Join(GetCwd(), "test_files")
	uploadData, err := os.ReadFile(filepath.Join(testFolder, "shrek.txt"))
	if err != nil {
		return 0, 0, fmt.Errorf("could not open test data: %v", err)
	}

	start2 := time.Now()

	_, err = client.Post(url, "application/octet-stream", bytes.NewReader(uploadData))
	if err != nil {
		return 0, 0, fmt.Errorf("could not upload test data: %v", err)
	}

	uploadDuration := time.Since(start2).Seconds()
	uploadSpeed = float64(len(uploadData)) / uploadDuration

	return downloadSpeed, uploadSpeed, nil
}

// This could be slow, depending on the users network. may also clog the network.
func averageSpeeds(iterations int, url string) (float64, float64) {
	var upTotal float64
	var downTotal float64

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	log.Print("[INFO] starting network profiling...")

	for i := 0; i < iterations; i++ {
		up, down, err := measureSpeed(url, client)
		if err != nil {
			log.Fatalf("[ERROR] error measuring average download and upload speed \n%v\n ", err)
		}
		upTotal += up
		downTotal += down
	}

	upAvg := upTotal / float64(iterations)
	downAvg := downTotal / float64(iterations)

	log.Printf("[INFO] up average: %f down average: %f\n", upAvg, downAvg)

	return upAvg, downAvg
}

// profile and save our average speeds as a network-profile.json file
// under ../sfs/pkg/network/profile/
func ProfileNetwork() *NetworkProfile {
	upAvg, dwnAvg := averageSpeeds(MAX, "www.google.com")

	profile := NewNetworkProfile()
	profile.UpRate = upAvg
	profile.DownRate = dwnAvg

	saveProfile(profile)

	return profile
}

// saveProfile creates a simple .json file of our network speed averages
func saveProfile(profile *NetworkProfile) error {
	// Open the JSON file for writing
	file, err := os.Create(filepath.Join(ProfileDirPath(), "network-profile.json"))
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
	if err = encoder.Encode(jsonStr); err != nil {
		log.Fatalf("[ERROR] error encoding network profile: %v", err)
	}

	return nil
}

/*
picks MAX limit for batch sizes based on the newly generated network profile

MAX = ((downrate * C) + (uprate * C)) / 2

current value for C is 0.75, and is totally arbitrary.
will probably fine tune/change equation as things develop.
*/
func PickMAX(p *NetworkProfile) int64 {
	return int64(((p.DownRate * 0.75) + (p.UpRate * 0.75)) / 2)
}
