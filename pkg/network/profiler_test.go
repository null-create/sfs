package network

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
)

func TestSaveProfile(t *testing.T) {
	path := ProfileDirPath()

	TmpProfile := NewNetworkProfile()
	saveProfile(TmpProfile)

	entries, err := os.ReadDir(path)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEqual(t, 0, len(entries))
	assert.Equal(t, 1, len(entries)) // should only have 1 file
	for _, entry := range entries {
		assert.True(t, strings.Contains(entry.Name(), "network-profile.json"))
	}

	if err := Clean(path); err != nil {
		t.Fatal(err)
	}
}

func TestMeasureSpeed(t *testing.T) {
	/*
		create a tmp folder to download a test file to.
		delete after testing

		start a small server to upload the test file from,
		then call measureSpeed() with the localhost:port url
		and download the test file from the tmp folder
		...to the tmp folder
	*/
	tmpDir := filepath.Join(GetCwd(), "tmp")
	testFile := filepath.Join(GetCwd(), "test_files/shrek.txt")

	err := os.Mkdir(tmpDir, 0644)
	if err != nil {
		Fatal(t, err)
	}

	// copy test file to tmp
	err = Copy(testFile, filepath.Join(tmpDir, "tmp.txt"))
	if err != nil {
		Fatal(t, err)
	}

	// define and start the server
	srv := http.Server{
		Handler:      http.FileServer(http.Dir(tmpDir)),
		Addr:         ":8080",
		ReadTimeout:  time.Second * 30,
		WriteTimeout: time.Second * 30,
		IdleTimeout:  time.Second * 30,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			Fatal(t, err)
		}
	}()

	// wait for server to start up
	log.Print("waiting for server to start...")
	time.Sleep(time.Second * 5)

	// "measure" speed
	log.Print("measuring speed...")
	down, up, err := measureSpeed("http://localhost:8080", &http.Client{
		Timeout: time.Second * 30,
	})
	if err != nil {
		Fatal(t, err)
	}
	assert.NotEqual(t, 0.0, down)
	assert.NotEqual(t, 0.0, up)
	// assert.True(t, down > 0.0)
	// assert.True(t, up > 0.0)

	// shut down server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := srv.Shutdown(ctx); err != nil {
		Fatal(t, err)
	}
	defer cancel()

	// clean up
	if err := Clean(tmpDir); err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(tmpDir); err != nil {
		t.Fatal(err)
	}
}

// func TestProfileNetwork(t *testing.T) {
// 	profile := ProfileNetwork()

// 	// clean up before our asserts so we don't leave anything behind
// 	if err := Clean(ProfileDirPath()); err != nil {
// 		t.Fatal(err)
// 	}

// 	assert.NotEqual(t, nil, profile)
// 	assert.NotEqual(t, "", profile.HostName)
// 	assert.NotEqual(t, 0, profile.UpRate)
// 	assert.NotEqual(t, 0, profile.DownRate)
// }
