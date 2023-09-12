package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi"
)

/*
Handlers for directly working with sfs service instance.

These will likely be called by middleware, which will themselves
be passed to the router when it is instantiated.

We want to add some middleware above these calls to handle user au
and other such business to validate requests to the server.
*/

type API struct {
	dbRoot string
}

func NewAPI() *API {
	c := ServiceConfig()
	return &API{
		dbRoot: filepath.Join(c.ServiceRoot, "dbs"),
	}
}

// -------- users -----------------------------------------

// attempts to read data from the user database.
//
// if found, it will attempt to prepare it as json data and return it
func (a *API) getUser(userID string) ([]byte, error) {
	u, err := findUser(userID, a.dbRoot)
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
		ServerErr(w, fmt.Sprintf("user (id=%s) not found", userID))
		return
	}
	w.Write(userData)
}

// -------- files -----------------------------------------

func (a *API) GetFileInfo(w http.ResponseWriter, r *http.Request) {
	fileID := chi.URLParam(r, "fileID")
	f, err := findFile(fileID, a.dbRoot)
	if err != nil {
		ServerErr(w, fmt.Sprintf("couldn't find file: %s", err.Error()))
		return
	}
	data, err := f.ToJSON()
	if err != nil {
		ServerErr(w, fmt.Sprintf("failed to convert to JSON: %s", err.Error()))
		return
	}
	w.Write(data)
}

// retrieve a file from the server
func (a *API) GetFile(w http.ResponseWriter, r *http.Request) {
	fileID := chi.URLParam(r, "fileID")
	f, err := findFile(fileID, a.dbRoot)
	if err != nil {
		ServerErr(w, err.Error())
		return
	}

	// Set the response header for the download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", f.Name))
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, f.ServerPath)
}

// upload or update a file on/to the server
func (a *API) PutFile(w http.ResponseWriter, r *http.Request) {
	fileID := chi.URLParam(r, "fileID")
	f, err := findFile(fileID, a.dbRoot)
	if err != nil {
		ServerErr(w, err.Error())
		return
	}

	// retrieve file from the request
	formFile, _, err := r.FormFile("myFile")
	if err != nil {
		ServerErr(w, fmt.Sprintf("failed to retrive form file data: %v", err))
		return
	}
	defer formFile.Close()

	// open (or create) the servers file for this user
	serverFile, err := os.Create(f.ServerPath)
	if err != nil {
		ServerErr(w, fmt.Sprintf("failed to create or truncate file: %v", err))
		return
	}
	defer serverFile.Close()

	// write file contents to server's pysical file
	_, err = io.Copy(serverFile, formFile)
	if err != nil {
		ServerErr(w, fmt.Sprintf("failed to create or truncate file: %v", err))
		return
	}
}

func (a *API) DeleteFile(w http.ResponseWriter, r *http.Request) {
	fileID := chi.URLParam(r, "fileID")
	f, err := findFile(fileID, a.dbRoot)
	if err != nil {
		ServerErr(w, err.Error())
		return
	}
	// remove physical file
	if err := os.Remove(f.ServerPath); err != nil {
		ServerErr(w, err.Error())
		return
	}
	// TODO: remove from svc instance... somehow
}

// ------- directories --------------------------------
