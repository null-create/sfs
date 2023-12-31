package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sfs/pkg/auth"
	svc "github.com/sfs/pkg/service"

	"github.com/go-chi/chi/v5"
)

/*
file for API implementations

these are the entry points into the SFS service via
http requests, and will have appropriate middleware calls
prior to most of these functions being called directly.
*/

type API struct {
	StartTime time.Time
	Svc       *Service // SFS service instance
}

func NewAPI(newService bool, isAdmin bool) *API {
	svc, err := Init(newService, isAdmin) // initialize sfs service
	if err != nil {
		log.Fatalf("[ERROR] failed to initialize new service instance: %v", err)
	}
	return &API{
		StartTime: time.Now().UTC(),
		Svc:       svc,
	}
}

// -------- testing ---------------------------------------

// placeholder handler for testing purposes
func (a *API) Placeholder(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("not implemented yet :("))
}

// -------- users (admin only) -----------------------------------------

// add a new user and drive to sfs instance. user existance and
// struct pointer should be created by NewUser middleware
func (a *API) AddNewUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(User).(*auth.User)
	if err := a.Svc.AddUser(user); err != nil {
		if strings.Contains(err.Error(), "user") {
			http.Error(w, err.Error(), http.StatusBadRequest) // user already exists
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		msg := fmt.Sprintf("user (name=%s id=%s) added", user.Name, user.ID)
		w.Write([]byte(msg))
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
	if u == nil {
		return nil, fmt.Errorf("user %s not found", userID)
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
		if strings.Contains(err.Error(), "not found") { // user not found
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	// TODO: userData should be part of a jwt token claim/payload,
	// and not just the bare bytes
	w.Write(userData)
}

// return a list of all active users
func (a *API) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if usrs, err := a.Svc.Db.GetUsers(); err == nil {
		if len(usrs) == 0 {
			w.Write([]byte("no users available"))
			return
		}
		for _, u := range usrs {
			data, err := u.ToJSON()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(data)
		}
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// use UserCtx middleware before a call to this
func (a *API) UpdateUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(User).(*auth.User)
	if err := a.Svc.UpdateUser(user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *API) DeleteUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(User).(*auth.User)
	if err := a.Svc.RemoveUser(user.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// -------- files -----------------------------------------

/*
NOTE: these are single operation file handlers.
syncing events will have separate handlers for file
uploads and downloads
*/

// check the db for the existence of a file.
//
// returns nil if file isn't found. handles db errors.
func (a *API) findF(w http.ResponseWriter, r *http.Request, fileID string) *svc.File {
	file, err := a.Svc.Db.GetFile(fileID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	return file
}

// get file metadata
func (a *API) GetFileInfo(w http.ResponseWriter, r *http.Request) {
	fileID := chi.URLParam(r, "fileID")
	if fileID == "" {
		http.Error(w, "missing file ID", http.StatusBadRequest)
		return
	}
	file := a.findF(w, r, fileID)
	if file == nil {
		http.Error(w, fmt.Sprintf("file (id=%s) not found", fileID), http.StatusNotFound)
		return
	}
	data, err := file.ToJSON()
	if err != nil {
		msg := fmt.Sprintf("failed to convert to JSON: %s", err.Error())
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// retrieve a file from the server
func (a *API) GetFile(w http.ResponseWriter, r *http.Request) {
	file := r.Context().Value(File).(*svc.File)

	// Set the response header for the download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name))
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, file.ServerPath)
}

// send an individual file to the server. shold be ran in its own goroutine.
func sendFile(w http.ResponseWriter, r *http.Request, file *svc.File) {
	// Set the response header for the download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name))
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, file.ServerPath)
}

// TODO: send multiple files to the client. each upload should happen in its own goroutine.
func (a *API) GetFiles(w http.ResponseWriter, r *http.Request) {
	// files := r.Context().Value(Files).([]*svc.File)
}

// get json blobs of all files available on the server.
// only sends metadata, not the actual files.
// TODO: implement a user-specific get-all-files db call
func (a *API) GetAllFiles(w http.ResponseWriter, r *http.Request) {
	if files, err := a.Svc.Db.GetFiles(); err == nil {
		if len(files) == 0 {
			w.Write([]byte("no files found"))
			return
		}
		for _, file := range files {
			data, err := file.ToJSON()
			if err != nil {
				msg := fmt.Sprintf("failed to convert file (name=%s id=%s) to JSON: %v", file.Name, file.ID, err)
				http.Error(w, msg, http.StatusInternalServerError)
				return
			}
			w.Write(data)
		}
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *API) newFile(w http.ResponseWriter, r *http.Request, newFile *svc.File) {
	// download the file
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r.Body)
	if err != nil {
		msg := fmt.Sprintf("failed to copy file data into buffer: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	if err := newFile.Save(buf.Bytes()); err != nil {
		msg := fmt.Sprintf("failed to save file to server: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	// update service. remove file if update fails.
	if err := a.Svc.AddFile(newFile.DirID, newFile); err != nil {
		if err := os.Remove(newFile.ServerPath); err != nil {
			log.Printf("[WARNING] failed to remove %s (id=%s) from server: %v", newFile.Name, newFile.ID, err)
		}
		msg := fmt.Sprintf("failed to add file to service: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	msg := fmt.Sprintf("%s has been added to the server", newFile.Name)
	log.Printf("[INFO] %s", msg)
	w.Write([]byte(msg))
}

// update the file
func (a *API) putFile(w http.ResponseWriter, r *http.Request, file *svc.File) {
	// retrieve file data from request body
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r.Body)
	if err != nil {
		msg := fmt.Sprintf("failed to download form file data: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	// update file
	if err := a.Svc.UpdateFile(file, buf.Bytes()); err != nil {
		http.Error(w, fmt.Sprintf("failed to update %s (id=%s)", file.Name, file.ID), http.StatusInternalServerError)
		return
	}
	msg := fmt.Sprintf("file %s updated", file.Name)
	log.Printf("[INFO] %s", msg)
	w.Write([]byte(msg))
}

// upload or update a file on/to the server
func (a *API) PutFile(w http.ResponseWriter, r *http.Request) {
	f := r.Context().Value(File).(*svc.File)
	if r.Method == http.MethodPut { // update the file
		a.putFile(w, r, f)
	} else if r.Method == http.MethodPost { // create a new file.
		a.newFile(w, r, f)
	}
}

// delete a file from the server
func (a *API) DeleteFile(w http.ResponseWriter, r *http.Request) {
	file := r.Context().Value(File).(*svc.File)
	if err := a.Svc.DeleteFile(file); err != nil {
		msg := fmt.Sprintf("failed to delete file: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	log.Printf("[INFO] file (%s) deleted from server", file.Path)
	w.Write([]byte(fmt.Sprintf("file (%s) deleted", file.Name)))
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

func (a *API) NewDir(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// TODO: refactor this to be a single context struct
	dirName := ctx.Value(Name).(string)
	parentID := ctx.Value(Parent).(string)
	driveID := ctx.Value(Drive).(string)
	path := ctx.Value(Path).(string)
	owner := ctx.Value(User).(string)

	// check if this directory exists
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		log.Printf("[INFO] dir %s already exists", path)
		http.Error(w, fmt.Sprintf("dir %s already exists", path), http.StatusBadRequest)
		return
	}
	// make the new physical directory & directory object
	if err := os.Mkdir(path, svc.PERMS); err != nil {
		log.Printf("[ERROR] failed to make directory: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dir := svc.NewDirectory(dirName, owner, driveID, path)

	// set parent to this directory
	parent, err := a.Svc.Db.GetDirectory(parentID)
	if err != nil {
		log.Printf("[ERROR] couldn't retrieve directory: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dir.Parent = parent

	// save new dir to directory db
	if err := a.Svc.Db.AddDir(dir); err != nil {
		log.Printf("[ERROR] failed to add directory to database: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("[INFO] dir %s (%s) created", dirName, path)

	msg := fmt.Sprintf("directory %s created", dirName)
	w.Write([]byte(msg))
}

func (a *API) DeleteDir(w http.ResponseWriter, r *http.Request) {

}

// -------- drives --------------------------------

func (a *API) GetDrive(w http.ResponseWriter, r *http.Request) {
	driveID := chi.URLParam(r, "driveID")
	if driveID == "" {
		http.Error(w, "missing drive ID", http.StatusBadRequest)
		return
	}
	// returns entire contents of the drive!
	drive := a.Svc.GetDrive(driveID)
	if drive == nil {
		http.Error(w, fmt.Sprintf("drive (id=%s) not found", driveID), http.StatusNotFound)
		return
	}
	data, err := drive.ToJSON()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to codify drive info to JSON: %v", err), http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// -------- sync ----------------------------------

func (a *API) Sync(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		driveID := chi.URLParam(r, "driveID")
		if driveID == "" {
			http.Error(w, "missing drive ID", http.StatusBadRequest)
			return
		}
		// attempt to get the sync index for this drive
		idx, err := a.Svc.GetSyncIdx(driveID)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to retrieve sync index: %v", err), http.StatusInternalServerError)
			return
		}
		if idx == nil { // no drive found
			http.Error(w, fmt.Sprintf("drive %s not found", driveID), http.StatusNotFound)
			return
		}
		data, err := idx.ToJSON()
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to encode sync index: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write(data)
	} else if r.Method == http.MethodPost {
		// TODO:
		w.Write([]byte("not implemented yet"))
	}
}
