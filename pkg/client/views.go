package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sfs/pkg/server"
)

var (
	homePage  = "http://" + cCfgs.Addr // web ui home page
	userPage  = homePage + "/user"     // users home page
	errorPage = homePage + "/error"
)

// -------- various pages -------------------------------------

func (c *Client) error(w http.ResponseWriter, r *http.Request, msg string, status int) {
	c.log.Error("Error: " + msg)
	errCtx := context.WithValue(r.Context(), server.Error, msg)
	http.Redirect(w, r.WithContext(errCtx), errorPage, http.StatusInternalServerError)
}

func (c *Client) HomePage(w http.ResponseWriter, r *http.Request) {
	recentDirs, err := c.Db.GetAllDirsAfter(time.Now().AddDate(0, 0, -7))
	if err != nil {
		c.error(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	recentFiles, err := c.Db.GetAllFilesAfter(time.Now().AddDate(0, 0, -7))
	if err != nil {
		c.error(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	indexData := Index{
		UserPage:   userPage,
		ProfilePic: c.Conf.ProfilePic,
		UserID:     c.User.ID,
		Files:      recentFiles,
		Dirs:       recentDirs,
		ServerHost: c.Conf.ServerAddr,
		ClientHost: c.Conf.Addr,
	}
	err = c.Templates.ExecuteTemplate(w, "index.html", indexData)
	if err != nil {
		c.error(w, r, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Client) DrivePage(w http.ResponseWriter, r *http.Request) {
	drivePageData := DrivePage{
		UserPage:   userPage,
		ProfilePic: c.Conf.ProfilePic,
		UserID:     c.User.ID,
		Files:      c.Drive.GetFiles(),
		Dirs:       c.Drive.GetDirs(),
		ServerHost: c.Conf.ServerAddr,
		ClientHost: c.Conf.Addr,
	}
	err := c.Templates.ExecuteTemplate(w, "drive.html", drivePageData)
	if err != nil {
		c.error(w, r, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Client) ErrorPage(w http.ResponseWriter, r *http.Request) {
	var errMsg string
	emsg := r.Context().Value(server.Error)
	if emsg == nil {
		errMsg = "he's dead, jim"
	} else {
		errMsg = emsg.(string)
	}
	errPageData := ErrorPage{
		UserPage:   userPage,
		ProfilePic: c.Conf.ProfilePic,
		ErrMsg:     errMsg,
	}
	err := c.Templates.ExecuteTemplate(w, "error.html", errPageData)
	if err != nil {
		c.error(w, r, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Client) UserPage(w http.ResponseWriter, r *http.Request) {
	usrPageData := UserPage{
		UserPage:   userPage,
		ProfilePic: c.Conf.ProfilePic,
		Name:       c.User.Name,
		UserID:     c.User.ID,
		UserName:   c.User.UserName,
		Email:      c.User.Email,
		TotalFiles: len(c.Drive.GetFiles()),
		TotalDirs:  len(c.Drive.GetDirs()),
		ServerHost: c.Conf.ServerAddr,
		ClientHost: c.Conf.Addr,
	}
	err := c.Templates.ExecuteTemplate(w, "user.html", usrPageData)
	if err != nil {
		c.error(w, r, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Client) RecycleBinPage(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(c.RecycleBin)
	if err != nil {
		c.error(w, r, err.Error(), http.StatusInternalServerError)
	}

	recycleBinItems := newRecycleBinItems()
	for _, entry := range entries {
		if entry.IsDir() {
			recycleBinItems.Dirs = append(recycleBinItems.Dirs, entry)
		} else {
			recycleBinItems.Files = append(recycleBinItems.Files, entry)
		}
	}

	recyclePageData := RecyclePage{
		UserPage:   userPage,
		ProfilePic: c.Conf.ProfilePic,
		Files:      recycleBinItems.Files,
		Dirs:       recycleBinItems.Dirs,
		ServerHost: c.Conf.ServerAddr,
		ClientHost: c.Conf.Addr,
	}
	err = c.Templates.ExecuteTemplate(w, "recycled.html", recyclePageData)
	if err != nil {
		c.error(w, r, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Client) DirPage(w http.ResponseWriter, r *http.Request) {
	dirID := r.Context().Value(server.Directory).(string)
	dir, err := c.GetDirectoryByID(dirID)
	if dir == nil {
		c.error(w, r, fmt.Sprintf("directory id=%s not found", dirID), http.StatusNotFound)
		return
	}
	if err != nil {
		c.error(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	// get sub directories
	subdirs, err := c.GetSubDirs(dirID)
	if err != nil {
		c.error(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	// get files
	files, err := c.GetFilesByDirID(dirID)
	if err != nil {
		c.error(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	dirPageData := DirPage{
		UserPage:   userPage,
		ProfilePic: c.Conf.ProfilePic,
		Name:       dir.Name,
		Size:       dir.Size,
		LastSync:   dir.LastSync,
		Dirs:       subdirs,
		Files:      files,
		ServerHost: c.Conf.ServerAddr,
		ClientHost: c.Conf.Addr,
	}
	err = c.Templates.ExecuteTemplate(w, "folder.html", dirPageData)
	if err != nil {
		c.error(w, r, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Client) FilePage(w http.ResponseWriter, r *http.Request) {
	fileID := r.Context().Value(server.File).(string)
	file, err := c.GetFileByID(fileID)
	if file == nil {
		http.Error(w, fmt.Sprintf("file (id=%s) not found", fileID), http.StatusNotFound)
		return
	}
	if err != nil {
		c.error(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	filePageData := FilePage{
		UserPage:   userPage,
		ProfilePic: c.Conf.ProfilePic,
		Name:       file.Name,
		Size:       file.Size,
		ID:         file.ID,
		OwnerID:    file.OwnerID,
		Type:       filepath.Ext(file.Name),
		LastSync:   file.LastSync,
		Location:   file.ClientPath,
		Checksum:   file.CheckSum,
		Endpoint:   file.Endpoint,
		ServerHost: c.Conf.ServerAddr,
		ClientHost: c.Conf.Addr,
	}
	err = c.Templates.ExecuteTemplate(w, "file.html", filePageData)
	if err != nil {
		c.error(w, r, err.Error(), http.StatusInternalServerError)
	}
}

// update client information received from the web ui
func (c *Client) EditInfo(w http.ResponseWriter, r *http.Request) {
	editPage := EditPage{
		UserPage:   userPage,
		ProfilePic: c.Conf.ProfilePic,
		Name:       c.User.Name,
		UserName:   c.User.UserName,
		Email:      c.Conf.Email,
		ServerHost: c.Conf.ServerAddr,
		ClientHost: c.Conf.Addr,
	}
	err := c.Templates.ExecuteTemplate(w, "edit.html", editPage)
	if err != nil {
		c.error(w, r, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Client) AddPage(w http.ResponseWriter, r *http.Request) {
	addPageData := AddPage{
		UserPage:     userPage,
		ProfilePic:   c.Conf.ProfilePic,
		DiscoverPath: "CHANGEME",
		ServerHost:   c.Conf.ServerAddr,
		ClientHost:   c.Conf.Addr,
		Endpoint:     c.Endpoints["new-file"],
	}
	err := c.Templates.ExecuteTemplate(w, "add.html", addPageData)
	if err != nil {
		c.error(w, r, err.Error(), http.StatusInternalServerError)
	}
}

// render upload page template
func (c *Client) UploadPage(w http.ResponseWriter, r *http.Request) {
	usersDirs, err := c.Db.GetUsersDirectories(c.UserID)
	if err != nil {
		c.error(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	uploadPageData := UploadPage{
		UserPage:   userPage,
		Dirs:       usersDirs,
		ProfilePic: c.Conf.ProfilePic,
		ServerHost: c.Conf.ServerAddr,
		ClientHost: c.Conf.Addr,
		Endpoint:   c.Endpoints["new file"],
	}
	err = c.Templates.ExecuteTemplate(w, "upload.html", uploadPageData)
	if err != nil {
		c.error(w, r, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Client) SettingsPage(w http.ResponseWriter, r *http.Request) {
	settingsPageData := SettingsPage{
		UserPage:        userPage,
		ServerHost:      c.Conf.ServerAddr,
		ClientHost:      c.Conf.Addr,
		ServerSync:      !c.Conf.ServerSync, // if local backup is disabed, we're syncing with the server
		BackupDir:       c.Conf.BackupDir,
		ClientPort:      c.Conf.ClientPort,
		EventBufferSize: c.Conf.EventBufferSize,
	}
	err := c.Templates.ExecuteTemplate(w, "settings.html", settingsPageData)
	if err != nil {
		c.error(w, r, err.Error(), http.StatusInternalServerError)
	}
}

// search for items
func (c *Client) SearchPage(w http.ResponseWriter, r *http.Request) {
	var resultsFile = "search-results.json"
	if r.Method == http.MethodPost {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r.Body)
		if err != nil {
			c.error(w, r, err.Error(), http.StatusInternalServerError)
			return
		}
		r.Body.Close()

		searchItem := buf.String()
		files, dirs, err := c.SearchForItems(searchItem)
		if err != nil {
			c.error(w, r, err.Error(), http.StatusInternalServerError)
			return
		}
		results := SearchResults{
			Files: files,
			Dirs:  dirs,
		}
		// save to temp json file so the search page
		// can display results when called with a GET request
		data, err := json.Marshal(results)
		if err != nil {
			c.error(w, r, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := os.WriteFile(resultsFile, data, 0644); err != nil {
			c.error(w, r, err.Error(), http.StatusInternalServerError)
			return
		}

	} else if r.Method == http.MethodGet {
		var results = NewSearchResults()
		if FileExists(resultsFile) {
			data, err := os.ReadFile(resultsFile)
			if err != nil {
				c.error(w, r, err.Error(), http.StatusInternalServerError)
				return
			}
			if err := json.Unmarshal(data, &results); err != nil {
				c.error(w, r, err.Error(), http.StatusInternalServerError)
				return
			}
			_ = os.Remove(resultsFile)
		}
		searchPageData := SearchPage{
			UserID:     c.UserID,
			UserPage:   userPage,
			ServerHost: c.Conf.ServerAddr,
			ClientHost: c.Conf.Addr,
			Files:      results.Files,
			Dirs:       results.Dirs,
		}
		err := c.Templates.ExecuteTemplate(w, "search.html", searchPageData)
		if err != nil {
			c.error(w, r, err.Error(), http.StatusInternalServerError)
		}
	}
}
