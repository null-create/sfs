package network

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	MAX       = 100
	URL       = "http://www.google.com"
	TEST_DATA = "/test_files/shrek.txt"
)

// We measure network resources by timing how long it takes to download a file of
// size N to URL, then how long it takes to upload data of the same size to said URL.
// This function is used iteratively to measure the average upload and download times,
// which helps the server determine file batch sizes
type NetworkProfile struct {
	HostName string  `json:"host"`
	UpRate   float64 `json:"up_rate"`
	DownRate float64 `json:"down_rate"`
}

func NetNetorkProfile() *NetworkProfile {
	return &NetworkProfile{
		HostName: GetHostName(),
	}
}

func measureSpeed(url string, client *http.Client) (downloadSpeed, uploadSpeed float64, err error) {
	//------Measure download speed
	start := time.Now()

	resp, err := client.Get(URL)
	if err != nil {
		return 0, 0, fmt.Errorf("[ERROR] failed to get URL: %v", err)
	}
	defer resp.Body.Close()

	downloadDuration := time.Since(start).Seconds()
	downloadSpeed = float64(resp.ContentLength) / downloadDuration

	//-----Measure upload speed with larger text file
	here, err := os.Getwd()
	if err != nil {
		return 0, 0, fmt.Errorf("[ERROR] failed to get current directory: %v", err)
	}
	uploadData, err := os.ReadFile(filepath.Join(here, TEST_DATA))
	if err != nil {
		return 0, 0, fmt.Errorf("[ERROR] could not open test data: %v", err)
	}

	start = time.Now()

	_, err = client.Post(URL, "application/octet-stream", bytes.NewReader(uploadData))
	if err != nil {
		return 0, 0, fmt.Errorf("[ERROR] could not upload test data: %v", err)
	}

	uploadDuration := time.Since(start).Seconds()
	uploadSpeed = float64(len(uploadData)) / uploadDuration

	return downloadSpeed, uploadSpeed, nil
}

// This could be slow, depending on the users network. may also clog the network.
func averageSpeeds(iterations int) (float64, float64) {
	var upTotal float64
	var downTotal float64

	log.Print("[DEBUG] starting network profiling...")

	for i := 0; i < iterations; i++ {
		up, down, err := measureSpeed(URL, &http.Client{
			Timeout: 10 * time.Second,
		})
		if err != nil {
			log.Fatalf("[ERROR] error measuring average download and upload speed \n%v\n ", err)
		}
		upTotal += up
		downTotal += down
	}

	upAvg := upTotal / float64(iterations)
	downAvg := downTotal / float64(iterations)

	log.Printf("[DEBUG] up average: %f down average: %f\n", upAvg, downAvg)

	return upAvg, downAvg
}

func ProfileNetwork() *NetworkProfile {
	profile := NetNetorkProfile()
	upAvg, dwnAvg := averageSpeeds(MAX)

	profile.UpRate = upAvg
	profile.DownRate = dwnAvg

	// save our average speeds as our network profile
	SaveProfile(profile)

	return profile
}
