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
	c := svc.ServiceConfig()
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

// TODO: refactor everything to use Service instance struct,
// and associated functions

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
		ServerErr(w, fmt.Sprintf("user (id=%s) not found", userID))
		return
	}
	w.Write(userData)
}

// -------- files -----------------------------------------

func (a *API) GetFileInfo(w http.ResponseWriter, r *http.Request) {
	fileID := chi.URLParam(r, "fileID")
	f, err := findFile(fileID, a.Db)
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
	f, err := findFile(fileID, a.Db)
	if err != nil {
		ServerErr(w, err.Error())
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
		ServerErr(w, fmt.Sprintf("failed to retrive form file data: %v", err))
	}
	defer formFile.Close()

	// TODO: get user's root directory using userID and by searching the DB for the path

	// retrieve file
	data := make([]byte, 0, header.Size)
	formFile.Read(data)

	// TODO: file integrity & safety checks. don't be stupid.
	// maybe file safety checks could be middleware

	// make file object & save to server under path
	fn := "change me"
	filePath := "change me"
	f := svc.NewFile(fn, userID, filePath)
	if err = f.Save(data); err != nil {
		ServerErr(w, err.Error())
	}
}

// save or update the file
func (a *API) putFile(w http.ResponseWriter, r *http.Request, f *svc.File) {
	formFile, header, err := r.FormFile("myFile")
	if err != nil {
		ServerErr(w, fmt.Sprintf("failed to retrive form file data: %v", err))
	}
	defer formFile.Close()

	data := make([]byte, header.Size)
	_, err = formFile.Read(data)
	if err != nil {
		ServerErr(w, err.Error())
	}
	if err := f.Save(data); err != nil {
		ServerErr(w, err.Error())
	}
}

// upload or update a file on/to the server
func (a *API) PutFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut { // update the file
		fileID := chi.URLParam(r, "fileID")
		f, err := findFile(fileID, a.Db)
		if err != nil {
			ServerErr(w, err.Error())
			return
		}
		a.putFile(w, r, f)
	} else if r.Method == http.MethodPost { // create a new file.
		userID := chi.URLParam(r, "userID")
		a.newFile(w, r, userID)
	}
}

func (a *API) DeleteFile(w http.ResponseWriter, r *http.Request) {
	fileID := chi.URLParam(r, "fileID")
	f, err := findFile(fileID, a.Db)
	if err != nil {
		ServerErr(w, err.Error())
		return
	}
	// remove physical file
	if err := os.Remove(f.ServerPath); err != nil {
		ServerErr(w, err.Error())
		return
	}
}

// ------- directories --------------------------------
