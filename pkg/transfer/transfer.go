package transfer

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"time"

	"github.com/sfs/pkg/auth"
	svc "github.com/sfs/pkg/service"
)

// transfer handles the uploading and downloading of individual files
// during synchronization events. one-off API calls are handled by
// other client request implementations.
type Transfer struct {
	Start  time.Time
	Buffer *bytes.Buffer
	Tok    *auth.Token
	Client *http.Client
}

func NewTransfer(port int) *Transfer {
	return &Transfer{
		Start:  time.Now().UTC(),
		Buffer: new(bytes.Buffer),
		Tok:    auth.NewT(),
		Client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

func (t *Transfer) dump(resp *http.Response, body bool) {
	b, err := httputil.DumpResponse(resp, body)
	if err != nil {
		log.Printf("[WARNING] failed to parse http response: %v", err)
	} else {
		log.Printf("[INFO] \n%v\n", string(b))
	}
}

func (t *Transfer) PrepareFileReq(method string, contentType string, file *svc.File, destURL string) (*http.Request, error) {
	req, err := http.NewRequest(method, destURL, t.Buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", contentType)

	// create file info token to attach to request header
	fileData, err := file.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to create file json string: %v", err)
	}
	fileToken, err := t.Tok.Create(string(fileData))
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

	fileWriter, err := bodyWriter.CreateFormFile("myFile", filepath.Base(file.Path))
	if err != nil {
		return err
	}
	if len(file.Content) == 0 {
		file.Load()
	}
	if _, err = fileWriter.Write(file.Content); err != nil {
		return fmt.Errorf("failed to retrieve file data: %v", err)
	}
	req, err := t.PrepareFileReq(method, bodyWriter.FormDataContentType(), file, file.Endpoint)
	if err != nil {
		return err
	}

	resp, err := t.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.dump(resp, true)
		return fmt.Errorf("server returned non-OK status: %v", resp.Status)
	}

	b, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Printf("[WARNING] failed to parse http response: %v", err)
	} else {
		log.Printf("[INFO] \n%v\n", string(b))
	}
	return nil
}

// download a known file from the given URL (associated server API endpoint).
//
// intended to run in its own goroutine.
// download a known file that is only on the server, and is new to the client
func (t *Transfer) Download(destPath string, srcURL string) error {
	resp, err := t.Client.Get(srcURL)
	if err != nil {
		return fmt.Errorf("failed to execute http request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.dump(resp, true)
		// server may be having issues.
		// does necessarily not mean a client error occurred.
		return nil
	}
	defer resp.Body.Close()

	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to copy file data to buffer: %v", err)
	}
	_, err = file.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write out file data: %v", err)
	}

	log.Printf("[INFO] %s downloaded to %s", file.Name(), destPath)
	return nil
}
