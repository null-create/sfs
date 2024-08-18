package client

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/sfs/pkg/logger"
	"github.com/sfs/pkg/server"
	svc "github.com/sfs/pkg/service"
)

// API handlers for the web client UI
// Must use http/template user of tmpl.Execute() instead of
// w.Write() whenever possible.

// generic response. sends msg with 200 and logs message.
func (c *Client) write(w http.ResponseWriter, msg string) {
	c.log.Log(logger.INFO, msg)
	w.Write([]byte(msg))
}

// not found response. sends a 404 and logs message.
func (c *Client) notFoundError(w http.ResponseWriter, err string) {
	c.log.Warn(err)
	http.Error(w, err, http.StatusNotFound)
}

// sends a bad request (400) with error message, and logs message
func (c *Client) clientError(w http.ResponseWriter, err string) {
	c.log.Warn(err)
	http.Error(w, err, http.StatusBadRequest)
}

// sends an internal server error (500) with an error message, and logs the message
func (c *Client) serverError(w http.ResponseWriter, err string) {
	c.log.Error(err)
	http.Error(w, err, http.StatusInternalServerError)
}

// Redirect will peform an HTTP redirect to the given redirect Path.
func (c *Client) Redirect(redirectPath string, req *http.Request) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, req, redirectPath, http.StatusFound)
	}
}

// enums mainly used for creating context keys
type Contexts string

const Error Contexts = "error"

// -------- various pages -------------------------------------

func (c *Client) HomePage(w http.ResponseWriter, r *http.Request) {
	indexData := Index{
		Files:      c.Drive.GetFiles(),
		Dirs:       c.Drive.GetDirs(),
		ServerHost: c.Conf.ServerAddr,
	}
	err := c.Templates.ExecuteTemplate(w, "index.html", indexData)
	if err != nil {
		c.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Client) ErrorPage(w http.ResponseWriter, r *http.Request) {
	errMsg := r.Context().Value(Error).(string)
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
	}
	err := c.Templates.ExecuteTemplate(w, "user.html", usrPageData)
	if err != nil {
		errCtx := context.WithValue(r.Context(), Error, err.Error())
		c.Redirect(homePage+"/error", r.WithContext(errCtx))
		return
	}
}

func (c *Client) DirPage(w http.ResponseWriter, r *http.Request) {
	dirID := r.Context().Value(server.Directory).(string)
	dir, err := c.GetDirectoryByID(dirID)
	if dir == nil {
		errCtx := context.WithValue(r.Context(), Error, fmt.Sprintf("dir id=%s not found", dirID))
		c.Redirect(homePage+"/error", r.WithContext(errCtx))
		return
	}
	if err != nil {
		errCtx := context.WithValue(r.Context(), Error, err.Error())
		c.Redirect(homePage+"/error", r.WithContext(errCtx))
		return
	}
	dirPageData := DirPage{
		Name:         dir.Name,
		Size:         dir.Size,
		TotalFiles:   len(dir.Files),
		TotalSubDirs: len(dir.Dirs),
		LastSync:     dir.LastSync,
		Files:        dir.GetFiles(),
		SubDirs:      dir.GetSubDirs(),
	}
	err = c.Templates.ExecuteTemplate(w, "folder.html", dirPageData)
	if err != nil {
		errCtx := context.WithValue(r.Context(), Error, err.Error())
		c.Redirect(homePage+"/error", r.WithContext(errCtx))
		return
	}
}

func (c *Client) FilePage(w http.ResponseWriter, r *http.Request) {
	fileID := r.Context().Value(server.File).(string)
	file, err := c.GetFileByID(fileID)
	if file == nil {
		errCtx := context.WithValue(r.Context(), Error, fmt.Sprintf("file id=%s not found", fileID))
		c.Redirect(homePage+"/error", r.WithContext(errCtx))
		return
	}
	if err != nil {
		errCtx := context.WithValue(r.Context(), Error, err.Error())
		c.Redirect(homePage+"/error", r.WithContext(errCtx))
		return
	}
	filePageData := FilePage{
		Name:     file.Name,
		Size:     file.Size,
		Type:     filepath.Ext(file.Name),
		Checksum: file.CheckSum,
		Endpoint: file.Endpoint,
		LastSync: file.LastSync,
	}
	err = c.Templates.ExecuteTemplate(w, "file.html", filePageData)
	if err != nil {
		errCtx := context.WithValue(r.Context(), Error, err.Error())
		c.Redirect(homePage+"/error", r.WithContext(errCtx))
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

func (c *Client) NewFile(w http.ResponseWriter, r *http.Request) {
	newFilePath := r.Context().Value(server.File).(string)
	if err := c.AddFile(newFilePath); err != nil {
		errCtx := context.WithValue(r.Context(), Error, err.Error())
		c.ErrorPage(w, r.WithContext(errCtx))
		return
	}
}

// retrieve a file from the server
func (c *Client) ServeFile(w http.ResponseWriter, r *http.Request) {
	f, err := c.getFileFromRequest(r)
	if err != nil {
		if strings.Contains(err.Error(), "file") { // not found or missing ID errors
			c.clientError(w, err.Error())
		} else {
			c.serverError(w, err.Error())
		}
		return
	}
	file, err := c.GetFileByID(f.ID)
	if err != nil {
		c.serverError(w, err.Error())
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name))
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, file.ClientPath)
	c.log.Info(fmt.Sprintf("served file %s: %s", file.Name, file.ClientPath))
}
