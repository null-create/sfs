package transfer

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"time"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/logger"
	svc "github.com/sfs/pkg/service"
)

// transfer handles the uploading and downloading of individual files
// during synchronization events as well as one off file transfer
// API calls.
type Transfer struct {
	Tok    *auth.Token
	log    *logger.Logger
	Client *http.Client
}

func NewTransfer() *Transfer {
	return &Transfer{
		Tok: auth.NewT(),
		log: logger.NewLogger("Transfer", "None"),
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
		t.log.Error("failed to parse http response: " + err.Error())
	} else {
		if resp.StatusCode == http.StatusOK {
			t.log.Log(logger.INFO, "server response: "+string(b))
		} else {
			t.log.Error(string(b))
		}
	}
}

// create a zip file of a directory so it can be transferred
func (t *Transfer) CreateArchive(path string) error {
	return Zip(path, path+".zip")
}

// extract contents of a zip file archive
func (t *Transfer) ExtractArchive(path string) error {
	return Unzip(path, filepath.Dir(path))
}

// prepare file transfer request header.
func (t *Transfer) PrepareFileReq(method string, destURL string, contentType string, file *svc.File, buf *bytes.Buffer) (*http.Request, error) {
	req, err := http.NewRequest(method, destURL, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// add file metadata to token
	fileData, err := file.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to create file json string: %v", err)
	}
	fileToken, err := t.Tok.Create(string(fileData))
	if err != nil {
		return nil, fmt.Errorf("failed to create file token: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+fileToken)
	req.Header.Set("Content-Type", contentType)

	return req, nil
}

// prepare and transfer a file for upload or download to the server.
// server will handle whether this is a new file or an update to an existing file,
// usually determined by the method.
func (t *Transfer) Upload(method string, file *svc.File, destURL string) error {
	var (
		buf = new(bytes.Buffer)
		w   = multipart.NewWriter(buf)
	)

	// create form file writer and prepare request
	fw, err := w.CreateFormFile("myFile", filepath.Base(file.Path))
	if err != nil {
		return err
	}

	// read in file data
	data, err := os.ReadFile(file.ClientPath)
	if err != nil {
		return err
	}
	if _, err = fw.Write(data); err != nil {
		return fmt.Errorf("failed to retrieve file data: %v", err)
	}
	if err := w.Close(); err != nil {
		t.log.Error("failed to close writer: " + err.Error())
	}

	// prepare request
	req, err := t.PrepareFileReq(method, destURL, w.FormDataContentType(), file, buf)
	if err != nil {
		return err
	}

	// send request
	t.log.Log("INFO", fmt.Sprintf("uploading %s to %s...", file.Name, file.Endpoint))
	resp, err := t.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %v", err)
	}
	t.dump(resp, true)
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

	// create (or truncate) file
	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer file.Close()

	// write out data
	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to copy file data to buffer: %v", err)
	}
	_, err = file.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write out file data: %v", err)
	}

	t.log.Log("INFO", fmt.Sprintf("%s downloaded to %s", file.Name(), destPath))
	return nil
}
