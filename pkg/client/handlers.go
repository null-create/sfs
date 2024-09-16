package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/sfs/pkg/server"
	svc "github.com/sfs/pkg/service"
)

// API handlers for the web client UI

const (
	DefaultSizeLimit = 1 << 10   // 1 kb default size limit for form data
	PhotoSizeLimit   = 10 << 20  // 10 mb file size limit for certain files (mostly profile pics)
	FileSizeLimit    = 100 << 30 // 100 gb file size limit (arbitrary size)
)

func (c *Client) successMsg(w http.ResponseWriter, msg string) {
	c.log.Info(msg)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": msg,
	})
}

// ------ users --------------------------------

func (c *Client) HandleNewUserInfo(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(DefaultSizeLimit)
	if err != nil {
		c.error(w, r, err.Error())
		return
	}
	// Extract form data (add more fields as needed)
	updates := map[string]string{
		"CLIENT_NAME":     r.FormValue("name"),
		"CLIENT_USERNAME": r.FormValue("username"),
		"CLIENT_EMAIL":    r.FormValue("email"),
	}

	// iterate through and if any are not empty and are different than
	// the current configurations, update in db and .env files accordingly
	for setting, newValue := range updates {
		if newValue != "" {
			if err := c.UpdateConfigSetting(setting, newValue); err != nil {
				c.error(w, r, err.Error())
				return
			}
		}
	}

	//success response
	c.successMsg(w, "settings updated successfully")
}

func (c *Client) UpdatePfpHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(PhotoSizeLimit) // Limit file size to 10MB
	if err != nil {
		http.Error(w, "Unable to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("profile-pic")
	if err != nil {
		http.Error(w, "Unable to retrieve file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileName := filepath.Base(header.Filename)
	dst, err := os.Create(fmt.Sprintf("./assets/profile-pics/%s", fileName))
	if err != nil {
		http.Error(w, "Unable to create the file on server: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Unable to save the file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update client configurations
	if err := c.UpdateConfigSetting("CLIENT_PROFILE_PIC", fileName); err != nil {
		c.error(w, r, err.Error())
		return
	}

	// Send back a JSON response with the new profile picture URL
	newProfilePicURL := fmt.Sprintf("/assets/profile-pics/%s", fileName)
	w.Header().Set("Content-Type", "application/json")
	_, err = fmt.Fprintf(w, `{"newProfilePicURL": "%s"}`, newProfilePicURL)
	if err != nil {
		http.Error(w, "Failed to format response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *Client) ClearPfpHandler(w http.ResponseWriter, r *http.Request) {
	if err := c.UpdateConfigSetting("CLIENT_PROFILE_PIC", "default_profile_pic.jpg"); err != nil {
		c.error(w, r, err.Error())
		return
	}
	c.successMsg(w, "profile pic updated successfully")
}

// empty the clients sfs recycle bin
func (c *Client) EmptyRecycleBinHandler(w http.ResponseWriter, r *http.Request) {
	if err := c.EmptyRecycleBin(); err != nil {
		c.error(w, r, err.Error())
		return
	}
	c.successMsg(w, "success")
}

// update config setting
func (c *Client) updateSetting(w http.ResponseWriter, setting string, value interface{}) {
	var v string
	if setting == "CLIENT_LOCAL_BACKUP" {
		v = strconv.FormatBool(!value.(bool)) // (server sync = false) == (client_local_backup = true)
	} else {
		v = value.(string)
	}
	if err := c.UpdateConfigSetting(setting, v); err != nil {
		c.log.Error("error updating settings: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// update client application settings from the web UI
func (c *Client) SettingsHandler(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r.Body)
	if err != nil {
		c.error(w, r, err.Error())
		return
	}
	r.Body.Close()

	var newSettings map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &newSettings); err != nil {
		c.error(w, r, err.Error())
		return
	}

	// update settings
	for setting, value := range newSettings {
		if setting != "" && value != "" {
			c.updateSetting(w, setting, value)
		}
	}

	c.successMsg(w, "settings updated successfully")
}

// -------- files -----------------------------------------

// gets a file from the ID provided by the request. returns nil if not found.
func (c *Client) getFileFromRequest(r *http.Request) (*svc.File, error) {
	fileID := r.Context().Value(server.File).(string)
	if fileID == "" {
		return nil, fmt.Errorf("no file ID found in request")
	}
	file, err := c.Db.GetFileByID(fileID)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, fmt.Errorf("file (id=%s) not found", fileID)
	}
	return file, nil
}

// open a file explorer window at the requested path
func (c *Client) OpenFileLocHandler(w http.ResponseWriter, r *http.Request) {
	file, err := c.getFileFromRequest(r)
	if err != nil {
		c.error(w, r, err.Error())
		return
	}
	if err := ShowFileInExplorer(file.ClientPath); err != nil {
		c.error(w, r, err.Error())
		return
	}
}

// retrieve a file from the local machine
func (c *Client) ServeFile(w http.ResponseWriter, r *http.Request) {
	file, err := c.getFileFromRequest(r)
	if err != nil {
		c.error(w, r, err.Error())
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name))
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, file.ClientPath)
	c.log.Info(fmt.Sprintf("served file %s: %s", file.Name, file.ClientPath))
}

func (c *Client) DropZoneHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(FileSizeLimit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	formFile, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer formFile.Close()

	destFolder := r.FormValue("destFolder")
	if destFolder == "" {
		http.Error(w, "Destination folder is required", http.StatusBadRequest)
		return
	}

	// see if this file is registered already
	file, err := c.Db.GetFileByName(handler.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if file != nil {
		w.Write([]byte(fmt.Sprintf("file '%s' already registered", handler.Filename)))
		return
	}

	// try to find the destination folder
	dir, err := c.Db.GetDirectoryByName(destFolder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if dir == nil {
		destFolder = c.Drive.Root.Path
	} else {
		destFolder = dir.Path
	}

	savePath := filepath.Join(destFolder, handler.Filename)
	c.log.Info(fmt.Sprintf("saving file to: %s", savePath))

	localFile, err := os.Create(savePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer localFile.Close()

	_, err = io.Copy(localFile, formFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// add uploaded file to service
	if err := c.AddFile(savePath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//success response
	c.successMsg(w, "file added successfully")
}

func (c *Client) RemoveFileHandler(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r.Body)
	if err != nil {
		c.error(w, r, err.Error())
		return
	}

	fileID := buf.String()
	if fileID == "" {
		http.Error(w, "no file ID provided", http.StatusBadRequest)
		return
	}

	file, err := c.GetFileByID(fileID)
	if err != nil {
		c.error(w, r, err.Error())
		return
	}
	if file == nil {
		c.error(w, r, "file not found")
		return
	}
	if err := c.RemoveFile(file); err != nil {
		c.error(w, r, err.Error())
		return
	}
}

// add a file or directory to the SFS service using its local path.
func (c *Client) AddItems(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r.Body)
	if err != nil {
		c.error(w, r, err.Error())
		return
	}
	r.Body.Close()

	path := buf.String()
	if path == "" {
		c.error(w, r, "no path provided")
		return
	}

	item, err := os.Stat(path)
	if err != nil {
		c.error(w, r, err.Error())
		return
	}

	if item.IsDir() {
		_, err = c.Discover(path)
		if err != nil {
			c.error(w, r, err.Error())
			return
		}
	} else {
		if err := c.AddFile(path); err != nil {
			c.error(w, r, err.Error())
			return
		}
	}

	c.successMsg(w, "items added successfully")
}
