package transfer

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// transfer handles the uploading and downloading of individual files.
// transfer operations are intended to run in their own goroutine as part
// of sync operations with the server
type Transfer struct {
	Start    time.Time
	Buffer   *bytes.Buffer
	Listener func(network string, address string) (net.Listener, error)
	Src      string // local file path of the file to be uploaded
	Dest     string // local destination for file downloads
	Client   *http.Client
}

func NewTransfer() *Transfer {
	return &Transfer{
		Start:    time.Now().UTC(),
		Buffer:   &bytes.Buffer{},
		Listener: net.Listen,
		Client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

func (t *Transfer) PrepareReq(contentType, destURL string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, destURL, t.Buffer)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] failed to create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", contentType)
	return req, nil
}

// prepare and transfer each file for upload or download
//
// intended to run within its own goroutine
func (t *Transfer) Upload(data []byte, fileName, destURL string) error {
	bodyWriter := multipart.NewWriter(t.Buffer)
	defer bodyWriter.Close()

	// create a form file field for the file
	fileWriter, err := bodyWriter.CreateFormFile("file", fileName)
	if err != nil {
		return err
	}

	// load data into file writer
	if _, err = fileWriter.Write(data); err != nil {
		return fmt.Errorf("[ERROR] failed to write file data: %v", err)
	}

	// prepare and send the request to the server
	req, err := t.PrepareReq(bodyWriter.FormDataContentType(), destURL)
	if err != nil {
		return err
	}
	log.Printf("[INFO] uploading %v ...", fileName)
	resp, err := t.Client.Do(req)
	if err != nil {
		return fmt.Errorf("[ERROR] failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[ERROR] server returned non-OK status: %v", resp.Status)
	}
	log.Printf("...done")

	return nil
}

// download a file from the given URL.
//
// intended to run in its own goroutine
func (t *Transfer) Download(dest, fileURL string) error {
	// listen for server requests
	ln, err := t.Listener("tcp", ":8080")
	if err != nil {
		return fmt.Errorf("[ERROR] failed to start client listener: %v", err)
	}
	defer ln.Close()

	// blocks until connection is established
	conn, err := ln.Accept()
	if err != nil {
		return fmt.Errorf("[ERROR] failed to create connection: %v", err)
	}
	defer conn.Close()

	// get file name & create local file
	fileNameBuffer := make([]byte, 0, 1024)
	n, err := conn.Read(fileNameBuffer)
	if err != nil {
		return fmt.Errorf("[ERROR] failed to read file name from server: %v", err)
	}
	fileName := string(fileNameBuffer[:n])
	file, err := os.Create(filepath.Join(dest, fileName))
	if err != nil {
		return fmt.Errorf("[ERROR] failed to create file: %v", err)
	}
	defer file.Close()

	// start downloading
	log.Printf("[INFO] downloading file %v ...", file)
	buffer := make([]byte, 0, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("[ERROR] failed to receive file data from server: %v", err)
			}
			break
		}
		_, err = file.Write(buffer[:n])
		if err != nil {
			return fmt.Errorf("[ERROR] failed to write file data: %v", err)
		}
	}
	log.Printf("...done")

	return nil
}
