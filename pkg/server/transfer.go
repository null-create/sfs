package server

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sfs/pkg/service"
)

// prepare and transfer each file for upload or download
//
// intended to run within its own goroutine
func Upload(data []byte, fileName string, destURL string) error {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// Create a form file field for the file
	fileWriter, err := bodyWriter.CreateFormFile("file", fileName)
	if err != nil {
		return fmt.Errorf("[ERROR] failed to create form file field: %v", err)
	}

	// load data into file writer
	if _, err = fileWriter.Write(data); err != nil {
		return fmt.Errorf("[ERROR] failed to write file data: %v", err)
	}

	// Close the multipart writer to finalize the form
	if err = bodyWriter.Close(); err != nil {
		return fmt.Errorf("[ERROR] failed to close multipart writer: %v", err)
	}

	// Make the HTTP request with the serialized file
	req, err := http.NewRequest("POST", destURL, bodyBuf)
	if err != nil {
		return fmt.Errorf("[ERROR] failed to create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("[ERROR] failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[ERROR] server returned non-OK status: %v", resp.Status)
	}

	return nil
}

// download a file from the given URL.
//
// intended to run in its own goroutine
func Download(fileURL string) (*service.File, error) {
	// TODO: designate specified ports at run times
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		return nil, fmt.Errorf("[ERROR] failed to start server: %v", err)
	}
	defer ln.Close()

	conn, err := ln.Accept()
	if err != nil {
		return nil, fmt.Errorf("[ERROR] failed to accept connection: %v", err)
	}
	defer conn.Close()

	fileNameBuffer := make([]byte, 1024)
	n, err := conn.Read(fileNameBuffer)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] failed to read file name from client: %v", err)
	}
	fileName := string(fileNameBuffer[:n])

	// TODO: add path
	savePath := filepath.Join("path/to/save/files", fileName)
	file, err := os.Create(savePath)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] failed to create file: %v", err)
	}
	defer file.Close()

	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				return nil, fmt.Errorf("[ERROR] failed to receive file data from client: %v", err)
			}
			break
		}

		_, err = file.Write(buffer[:n])
		if err != nil {
			return nil, fmt.Errorf("[ERROR] failed to write file data: %v", err)
		}
	}

	return nil, nil
}
