package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi"

	"github.com/sfs/pkg/db"
	svc "github.com/sfs/pkg/service"
)

/*
Handlers for directly working with sfs service instance.

These will likely be called by middleware, which will themselves
be passed to the router when it is instantiated.

We want to add some middleware above these calls to handle user au
and other such business to validate requests to the server.
*/

type API struct {
	Db  *db.Query // db connection
	Svc *Service  // SFS service instance
}

func NewAPI(newService bool, isAdmin bool) *API {
	// get service config and db connection
	c := ServiceConfig()
	db := db.NewQuery(filepath.Join(c.S.SvcRoot, "dbs"), true)

	// initialize sfs service
	svc, err := Init(newService, isAdmin)
	if err != nil {
		log.Fatalf("[ERROR] failed to initialize new service instance: %v", err)
	}
	return &API{
		Db:  db,
		Svc: svc,
	}
}

// -------- testing ---------------------------------------

// placeholder handler for testing purposes
func (a *API) Placeholder(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hi"))
}

// -------- users -----------------------------------------

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

// -------- files -----------------------------------------

// check the db for the existence of a file.
//
// handles errors and not found cases. returns nil if either of these
// are the case, otherwise returns a file pointer.
func (a *API) findF(w http.ResponseWriter, r *http.Request) *svc.File {
	fileID := chi.URLParam(r, "fileID")
	f, err := findFile(fileID, a.Db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	} else if f == nil {
		msg := fmt.Sprintf("file (id=%s) not found", fileID)
		http.Error(w, msg, http.StatusNotFound)
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

	// TODO: get user's root directory using userID and by searching the DB for the path

	// retrieve file
	data := make([]byte, 0, header.Size)
	formFile.Read(data)

	// TODO: file integrity & safety checks. don't be stupid.

	// make file object & save to server under path
	fn := "change me"
	filePath := "change me"
	f := svc.NewFile(fn, userID, filePath)
	if err = f.Save(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// save or update the file
func (a *API) putFile(w http.ResponseWriter, r *http.Request, f *svc.File) {
	formFile, header, err := r.FormFile("myFile")
	if err != nil {
		msg := fmt.Sprintf("failed to retrive form file data: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	defer formFile.Close()

	data := make([]byte, 0, header.Size)
	_, err = formFile.Read(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if err := f.Save(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		ServerErr(w, err.Error())
	}
	// TODO: remove from db and maybe user instance?
}

// ------- directories --------------------------------

// check the db for the existence of a directory.
//
// handles errors and not found cases. returns nil if either of these,
// are the case, otherwise returns a directory pointer.
func (a *API) findD(w http.ResponseWriter, r *http.Request) *svc.Directory {
	dirID := chi.URLParam(r, "dirID")
	d, err := findDir(dirID, a.Db)
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
