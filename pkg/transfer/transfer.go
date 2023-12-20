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
	"net/http/httputil"
	"os"
	"path/filepath"
	"time"

	"github.com/sfs/pkg/auth"
	svc "github.com/sfs/pkg/service"
)

// transfer handles the uploading and downloading of individual files.
// transfer operations are intended to run in their own goroutine as part
// of sync operations with the server
type Transfer struct {
	Start  time.Time
	Buffer *bytes.Buffer

	// dedicated listener for downloads
	Listener func(network string, address string) (net.Listener, error)
	Port     int // port to listen to for downloads

	Client *http.Client
}

func NewTransfer(port int) *Transfer {
	return &Transfer{
		Start:    time.Now().UTC(),
		Buffer:   new(bytes.Buffer),
		Listener: net.Listen,
		Port:     port,
		Client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

func (t *Transfer) PrepareFileReq(method string, contentType string, file *svc.File, destURL string) (*http.Request, error) {
	req, err := http.NewRequest(method, destURL, t.Buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", contentType)

	// create file info token to attach to request header
	tokenizer := auth.NewT()
	fileData, err := file.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to create file json string: %v", err)
	}
	fileToken, err := tokenizer.Create(string(fileData))
	if err != nil {
		return nil, fmt.Errorf("failed to create file token: %v", err)
	}
	req.Header.Set("Authorization", fileToken)

	return req, nil
}

// prepare and transfer a file for upload or download to the server.
// server will handle whether this is a new file or an update to an existing file,
// usually determined by the method.
//
// intended to run within its own goroutine.
func (t *Transfer) Upload(method string, file *svc.File, destURL string) error {
	bodyWriter := multipart.NewWriter(t.Buffer)
	defer bodyWriter.Close()

	// create a form file field for the file
	fileWriter, err := bodyWriter.CreateFormFile("myFile", filepath.Base(file.Path))
	if err != nil {
		return err
	}
	// load data into file writer, then prepare and send the request to the destination
	if len(file.Content) == 0 {
		file.Load()
	}
	if _, err = fileWriter.Write(file.Content); err != nil {
		return fmt.Errorf("failed to retrieve file data: %v", err)
	}
	req, err := t.PrepareFileReq(method, bodyWriter.FormDataContentType(), file, destURL)
	if err != nil {
		return err
	}

	// upload and confirm success
	log.Printf("[INFO] uploading %v ...", filepath.Base(file.Path))
	resp, err := t.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("[WARNING] server returned non-OK status: %v", resp.Status)
		b, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Printf("[WARNING] failed to parse http response: %v", err)
			return nil
		}
		return fmt.Errorf("failed to upload file: %v", fmt.Sprintf("\n%s\n", string(b)))
	}

	log.Printf("[INFO] ...done")
	b, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Printf("[WARNING] failed to parse http response: %v", err)
		return nil
	}
	log.Printf("[INFO] \n%v\n", string(b))
	return nil
}

// download a known file from the given URL (associated server API endpoint).
//
// intended to run in its own goroutine.
// download a known file that is only on the server, and is new to the client
func (t *Transfer) Download(destPath string, fileURL string) error {
	// attempt to retrieve the file from the server
	resp, err := t.Client.Get(fileURL)
	if err != nil {
		return fmt.Errorf("failed to execute http request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		b, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Printf("[WARNING] failed to parse response: %v", err)
		} else {
			log.Printf("[INFO] failed to retrieve file: %v", string(b))
		}
		return nil
	}
	defer resp.Body.Close()

	// create destination file
	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to download file: %v", err)
	}
	_, err = file.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to download file: %v", err)
	}

	log.Printf("[INFO] file downloaded")
	return nil
}
