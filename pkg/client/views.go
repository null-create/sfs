package client

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sfs/pkg/server"
)

var (
	homePage  = "http://" + cfgs.Addr // web ui home page
	userPage  = homePage + "/user"    // users home page
	errorPage = homePage + "/error"
)

// -------- various pages -------------------------------------

func (c *Client) error(w http.ResponseWriter, r *http.Request, msg string) {
	c.log.Error("Error: " + msg)
	errCtx := context.WithValue(r.Context(), server.Error, msg)
	http.Redirect(w, r.WithContext(errCtx), errorPage, http.StatusInternalServerError)
}

func (c *Client) HomePage(w http.ResponseWriter, r *http.Request) {
	indexData := Index{
		UserPage:   userPage,
		ProfilePic: c.Conf.ProfilePic,
		UserID:     c.User.ID,
		Files:      c.Drive.GetFiles(),
		Dirs:       c.Drive.GetDirs(),
		ServerHost: c.Conf.ServerAddr,
		ClientHost: c.Conf.Addr,
	}
	err := c.Templates.ExecuteTemplate(w, "index.html", indexData)
	if err != nil {
		c.error(w, r, err.Error())
	}
}

func (c *Client) ErrorPage(w http.ResponseWriter, r *http.Request) {
	errMsg := r.Context().Value(server.Error)
	if errMsg == nil {
		http.Error(w, "No error parsed from request", http.StatusInternalServerError)
		return
	}
	errMsg = errMsg.(string)
	errPageData := ErrorPage{
		UserPage:   userPage,
		ProfilePic: c.Conf.ProfilePic,
		ErrMsg:     fmt.Sprintf("Something went wrong :(\n\n%s", errMsg),
	}
	err := c.Templates.ExecuteTemplate(w, "error.html", errPageData)
	if err != nil {
		c.error(w, r, err.Error())
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
		c.error(w, r, err.Error())
	}
}

func (c *Client) RecycleBinPage(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(c.RecycleBin)
	if err != nil {
		c.error(w, r, err.Error())
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
		c.error(w, r, err.Error())
	}
}

func (c *Client) DirPage(w http.ResponseWriter, r *http.Request) {
	dirID := r.Context().Value(server.Directory).(string)
	dir, err := c.GetDirectoryByID(dirID)
	if dir == nil {
		c.error(w, r, err.Error())
		return
	}
	if err != nil {
		c.error(w, r, err.Error())
		return
	}
	// get sub directories
	subdirs, err := c.GetSubDirs(dirID)
	if err != nil {
		c.error(w, r, err.Error())
		return
	}
	// get files
	files, err := c.GetFilesByDirID(dirID)
	if err != nil {
		c.error(w, r, err.Error())
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
		c.error(w, r, err.Error())
	}
}

func (c *Client) FilePage(w http.ResponseWriter, r *http.Request) {
	fileID := r.Context().Value(server.File).(string)
	file, err := c.GetFileByID(fileID)
	if file == nil {
		c.error(w, r, err.Error())
		return
	}
	if err != nil {
		c.error(w, r, err.Error())
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
		c.error(w, r, err.Error())
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
		c.error(w, r, err.Error())
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
		c.error(w, r, err.Error())
	}
}
