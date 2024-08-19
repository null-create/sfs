package client

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/sfs/pkg/server"
	svc "github.com/sfs/pkg/service"
)

// API handlers for the web client UI

// enums mainly used for creating context keys
type Contexts string

const Error Contexts = "error"

// Redirect will peform an HTTP redirect to the given redirect Path.
func (c *Client) Redirect(redirectPath string, req *http.Request) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, req, redirectPath, http.StatusFound)
	}
}

// -------- various pages -------------------------------------

func (c *Client) HomePage(w http.ResponseWriter, r *http.Request) {
	indexData := Index{
		Files:      c.Drive.GetFiles(),
		Dirs:       c.Drive.GetDirs(),
		ServerHost: c.Conf.ServerAddr,
		ClientHost: c.Conf.Addr,
	}
	err := c.Templates.ExecuteTemplate(w, "index.html", indexData)
	if err != nil {
		http.Redirect(w, r, homePage+"/error/"+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *Client) ErrorPage(w http.ResponseWriter, r *http.Request) {
	// errMsg := r.Context().Value(Error).(string)
	errMsg := "oops"
	errPageData := ErrorPage{
		ErrMsg: fmt.Sprintf("Something went wrong :(\n\n%s", errMsg),
	}
	err := c.Templates.ExecuteTemplate(w, "error.html", errPageData)
	if err != nil {
		c.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Client) UserPage(w http.ResponseWriter, r *http.Request) {
	usrPageData := UserPage{
		Name:           c.User.Name,
		UserName:       c.User.UserName,
		Email:          c.User.Email,
		TotalFiles:     len(c.Drive.GetFiles()),
		TotalDirs:      len(c.Drive.GetDirs()),
		ProfilePicPath: "",
		ServerHost:     c.Conf.ServerAddr,
		ClientHost:     c.Conf.Addr,
	}
	err := c.Templates.ExecuteTemplate(w, "user.html", usrPageData)
	if err != nil {
		http.Redirect(w, r, homePage+"/error/"+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *Client) DirPage(w http.ResponseWriter, r *http.Request) {
	dirID := r.Context().Value(server.Directory).(string)
	dir, err := c.GetDirectoryByID(dirID)
	if dir == nil {
		http.Redirect(w, r, homePage+"/error/"+err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Redirect(w, r, homePage+"/error/"+err.Error(), http.StatusInternalServerError)
		return
	}
	// get sub directories
	subdirs, err := c.GetSubDirs(dirID)
	if err != nil {
		c.Redirect(homePage+"/error/"+err.Error(), r)
		return
	}
	// get files
	files, err := c.GetFilesByDirID(dirID)
	if err != nil {
		http.Redirect(w, r, homePage+"/error/"+err.Error(), http.StatusInternalServerError)
		return
	}
	dirPageData := DirPage{
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
		http.Redirect(w, r, homePage+"/error/"+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *Client) FilePage(w http.ResponseWriter, r *http.Request) {
	fileID := r.Context().Value(server.File).(string)
	file, err := c.GetFileByID(fileID)
	if file == nil {
		http.Redirect(w, r, homePage+"/error/"+err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Redirect(w, r, homePage+"/error/"+err.Error(), http.StatusInternalServerError)
		return
	}
	filePageData := FilePage{
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
		c.Redirect(homePage+"/error/"+err.Error(), r)
		return
	}
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

// add a new file to the SFS serverice using its absolute path
func (c *Client) NewFile(w http.ResponseWriter, r *http.Request) {
	newFilePath := r.Context().Value(server.File).(string)
	if err := c.AddFile(newFilePath); err != nil {
		http.Redirect(w, r, homePage+"/error/"+err.Error(), http.StatusInternalServerError)
		return
	}
}

// retrieve a file from the local machine
func (c *Client) ServeFile(w http.ResponseWriter, r *http.Request) {
	file, err := c.getFileFromRequest(r)
	if err != nil {
		http.Redirect(w, r, homePage+"/error/"+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name))
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, file.ClientPath)
	c.log.Info(fmt.Sprintf("served file %s: %s", file.Name, file.ClientPath))
}
