package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/sfs/pkg/auth"
	svc "github.com/sfs/pkg/service"

	"github.com/go-chi/chi"
)

/*
Handlers for directly working with sfs service instance..

We want to add some middleware above these calls to handle user authorization
and other such business to validate requests to the server.
*/

type API struct {
	Svc *Service // SFS service instance
}

func NewAPI(newService bool, isAdmin bool) *API {
	svc, err := Init(newService, isAdmin) // initialize sfs service
	if err != nil {
		log.Fatalf("[ERROR] failed to initialize new service instance: %v", err)
	}
	return &API{
		Svc: svc,
	}
}

// -------- testing ---------------------------------------

// placeholder handler for testing purposes
func (a *API) Placeholder(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("not implemented yet :("))
}

// -------- users (admin only) -----------------------------------------

// generate new user instance, and create drive and other base files
func (a *API) AddUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	if err := a.Svc.AddUser(userID, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// attempts to read data from the user database.
//
// if found, it will attempt to prepare it as json data and return it
func (a *API) getUser(userID string) ([]byte, error) {
	u, err := a.Svc.FindUser(userID)
	if err != nil {
		return nil, err
	}
	jsonData, err := u.ToJSON()
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func (a *API) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	userData, err := a.getUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(userData)
}

// TODO: figure out how to determine which user data is new, and
// how to retrieve the new data from the request object
func (a *API) updateUser(user *auth.User, r *http.Request) error {

	return nil
}

func (a *API) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	u, err := a.Svc.FindUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if u == nil {
		http.Error(w, fmt.Sprintf("user %s not found", userID), http.StatusNotFound)
		return
	}
	if err := a.updateUser(u, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *API) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	u, err := a.Svc.FindUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := a.Svc.RemoveUser(u.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// -------- files -----------------------------------------

// check the db for the existence of a file.
//
// handles errors and not found cases. returns nil if either of these
// are the case, otherwise returns a file pointer.
func (a *API) findF(w http.ResponseWriter, r *http.Request) *svc.File {
	fileID := chi.URLParam(r, "fileID")
	f, err := a.Svc.Db.GetFile(fileID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	} else if f == nil {
		msg := fmt.Sprintf("file (id=%s) not found", fileID)
		http.Error(w, msg, http.StatusNotFound)
		return nil
	}
	return f
}

// get file metadata
func (a *API) GetFileInfo(w http.ResponseWriter, r *http.Request) {
	f := a.findF(w, r)
	if f == nil {
		return
	}
	data, err := f.ToJSON()
	if err != nil {
		msg := fmt.Sprintf("failed to convert to JSON: %s", err.Error())
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// retrieve a file from the server
func (a *API) GetFile(w http.ResponseWriter, r *http.Request) {
	f := a.findF(w, r)
	if f == nil {
		return
	}

	// Set the response header for the download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", f.Name))
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, f.ServerPath)
}

func (a *API) newFile(w http.ResponseWriter, r *http.Request, userID string) {
	formFile, header, err := r.FormFile("myFile")
	if err != nil {
		msg := fmt.Sprintf("failed to retrive form file data: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	defer formFile.Close()

	// TODO: file integrity & safety checks. don't be stupid.

	// get the user associated with the file
	u, err := a.Svc.Db.GetUser(userID)
	if err != nil {
		msg := fmt.Sprintf("failed to get user from db: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	// retrieve file
	var buf bytes.Buffer
	_, err = io.Copy(&buf, formFile)
	if err != nil {
		msg := fmt.Sprintf("failed to retrive form file data: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	// make file object & save to server under path
	filePath := "change me"
	f := svc.NewFile(header.Filename, u.Name, filePath)

	if err := f.Save(buf.Bytes()); err != nil {
		msg := fmt.Sprintf("failed to download file to server: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	f.LastSync = time.Now().UTC()

	a.Svc.Db.WhichDB("files")
	if err := a.Svc.Db.AddFile(f); err != nil {
		msg := fmt.Sprintf("failed to add file to database: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
}

// update the file
func (a *API) putFile(w http.ResponseWriter, r *http.Request, f *svc.File) {
	formFile, _, err := r.FormFile("myFile")
	if err != nil {
		msg := fmt.Sprintf("failed to retrive form file data: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	defer formFile.Close()

	// retrieve file & update last sync time
	var buf bytes.Buffer
	_, err = io.Copy(&buf, formFile)
	if err != nil {
		msg := fmt.Sprintf("failed to download form file data: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	if err := f.Save(buf.Bytes()); err != nil {
		msg := fmt.Sprintf("failed to download file to server: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	f.LastSync = time.Now().UTC()

	// update DB
	a.Svc.Db.WhichDB("users")
	if err := a.Svc.Db.UpdateFile(f); err != nil {
		msg := fmt.Sprintf("failed to update file database: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
}

// upload or update a file on/to the server
func (a *API) PutFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut { // update the file
		f := a.findF(w, r)
		if f == nil {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		a.putFile(w, r, f)
	} else if r.Method == http.MethodPost { // create a new file.
		userID := chi.URLParam(r, "userID")
		// TODO: get destination file path ...somehow.
		// should be a parameter to newFile()
		a.newFile(w, r, userID)
	}
}

// delete a file from the server
func (a *API) DeleteFile(w http.ResponseWriter, r *http.Request) {
	f := a.findF(w, r)
	if f == nil {
		return
	}
	// remove physical file
	if err := os.Remove(f.ServerPath); err != nil {
		msg := fmt.Sprintf("failed to remove file from server: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	// remove from database
	if err := a.Svc.Db.RemoveFile(f.ID); err != nil {
		msg := fmt.Sprintf("failed to remove file from database: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
}

// ------- directories --------------------------------

// check the db for the existence of a directory.
//
// handles errors and not found cases. returns nil if either of these
// are the case, otherwise returns a directory pointer.
func (a *API) findD(w http.ResponseWriter, r *http.Request) *svc.Directory {
	dirID := chi.URLParam(r, "dirID")
	d, err := findDir(dirID, a.Svc.Db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	} else if d == nil {
		msg := fmt.Sprintf("directory (id=%s) not found: ", err.Error())
		http.Error(w, msg, http.StatusNotFound)
		return nil
	}
	return d
}

func (a *API) GetDirectory(w http.ResponseWriter, r *http.Request) {
	d := a.findD(w, r)
	if d == nil {
		return
	}
	data, err := d.ToJSON()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
