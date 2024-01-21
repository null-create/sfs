package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sfs/pkg/auth"
	svc "github.com/sfs/pkg/service"
	"github.com/sfs/pkg/transfer"
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
	w.Write([]byte("\nnot implemented yet :(\n"))
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
		msg := fmt.Sprintf("\nuser (name=%s id=%s) added\n", user.Name, user.ID)
		w.Write([]byte(msg))
	}
}

// send user metadata.
func (a *API) GetUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(User).(*auth.User)
	userData, err := user.ToJSON()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// TODO: more secure way to send user data.
	// this just sends the raw JSON response.
	w.Write(userData)
}

// return a list of all active users. used for admin and testing purposes.
func (a *API) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users := r.Context().Value(Users).([]*auth.User)
	for _, u := range users {
		data, err := u.ToJSON()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(data)
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

// remove a user from the server
func (a *API) DeleteUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(User).(*auth.User)
	if err := a.Svc.RemoveUser(user.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// -------- files -----------------------------------------

// get file metadata
func (a *API) GetFileInfo(w http.ResponseWriter, r *http.Request) {
	file := r.Context().Value(File).(*svc.File)
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
	// file existance as confirmed in middleware by this point
	file := r.Context().Value(File).(*svc.File)

	// Set the response header for the download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name))
	w.Header().Set("Content-Type", "application/octet-stream")

	// send the file
	http.ServeFile(w, r, file.ServerPath)
}

// get json blobs of all files available on the server for a user.
// only sends metadata, not the actual files.
func (a *API) GetAllFileInfo(w http.ResponseWriter, r *http.Request) {
	files := r.Context().Value(Files).([]*svc.File)
	for _, file := range files {
		data, err := file.ToJSON()
		if err != nil {
			msg := fmt.Sprintf("\nfailed to convert %s (id=%s) to JSON: %v\n", file.Name, file.ID, err)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		w.Write(data)
	}
}

func (a *API) newFile(w http.ResponseWriter, r *http.Request, newFile *svc.File) {
	// download the file
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r.Body)
	if err != nil {
		msg := fmt.Sprintf("\nfailed to copy file data into buffer: %v\n", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	// save file to server
	if err := newFile.Save(buf.Bytes()); err != nil {
		msg := fmt.Sprintf("\nfailed to save file to server: %v\n", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	// update service
	if err := a.Svc.AddFile(newFile.DirID, newFile); err != nil {
		http.Error(w, fmt.Sprintf("\nfailed to add file to service: %v\n", err), http.StatusInternalServerError)
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
		http.Error(w, fmt.Sprintf("\nfailed to update %s (id=%s)\n", file.Name, file.ID), http.StatusInternalServerError)
		return
	}
	msg := fmt.Sprintf("\n%s updated (owner id=%s)\n", file.Name, file.OwnerID)
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
		http.Error(w, fmt.Sprintf("\nfailed to delete file: %v\n", err), http.StatusInternalServerError)
		return
	}
	msg := fmt.Sprintf("\n%s (id=%s) deleted from server\n", file.Name, file.ID)
	log.Printf("[INFO] %s ", msg)
	w.Write([]byte(msg))
}

// ------- directories --------------------------------

// returns metadata for a single directory (not its children).
func (a *API) GetDirInfo(w http.ResponseWriter, r *http.Request) {
	dir := r.Context().Value(Directory).(*svc.Directory)
	data, err := dir.ToJSON()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func (a *API) walkDir(w http.ResponseWriter, dir *svc.Directory) error {
	// send file info, if any
	for _, file := range dir.Files {
		fileData, err := file.ToJSON()
		if err != nil {
			return err
		}
		w.Write(fileData)
	}
	if len(dir.Dirs) == 0 {
		return nil
	}
	// send any subdirectory info before descending directory tree further
	for _, subDir := range dir.Dirs {
		sdData, err := subDir.ToJSON()
		if err != nil {
			return err
		}
		w.Write(sdData)
	}
	for _, subDir := range dir.Dirs {
		if err := a.walkDir(w, subDir); err != nil {
			log.Printf("[WARNING] %v", err)
		}
	}
	return nil
}

// retrieve metadata for a directory as well as all its files and children.
// does not return file contents, only metadata.
func (a *API) GetManyDirsInfo(w http.ResponseWriter, r *http.Request) {
	dir := r.Context().Value(Directory).(*svc.Directory)
	dir = a.Svc.Populate(dir)
	// walk the directory tree starting from this directory, and
	// send JSON blobs for each object it discovers along the way.
	// data is sent in depth-first search order so client side
	// will need to sort that out.
	a.walkDir(w, dir)
}

// retrieve a zipfile of the directory (and all its children)
func (a *API) GetDir(w http.ResponseWriter, r *http.Request) {
	dir := r.Context().Value(Directory).(*svc.Directory)
	// create a tmp .zip file so we can transfer the directory and its contents
	archive := filepath.Join(dir.Path, fmt.Sprintf(dir.Name, ".zip"))
	if err := transfer.Zip(dir.Path, archive); err != nil {
		http.Error(w, fmt.Sprintf("\nfailed to compress directory: %v\n", err), http.StatusInternalServerError)
		return
	}
	// send archive file
	http.ServeFile(w, r, archive)
	// remove tmp .zip file and tmp folder
	if err := os.Remove(archive); err != nil {
		log.Printf("[WARNING] failed to remove temp archive %s: %v", archive, err)
	}
}

// TODO:
// create a new directory with supplied contents on the server.
// should take a .zip file sent from the user, unpack it in the
// desired location, and update internal data structures and databases
// accordingly

// func (a *API) PutDir(w http.ResponseWriter, r *http.Request) {
//
// }

// create a new empty physical directory on the server for a user
func (a *API) NewDir(w http.ResponseWriter, r *http.Request) {
	newDir := r.Context().Value(Directory).(*svc.Directory)
	if err := a.Svc.NewDir(newDir.DriveID, newDir.Parent.ID, newDir); err != nil {
		http.Error(w, fmt.Sprintf("failed to create directory: %v", err), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(fmt.Sprintf("\ndirectory %s (id=%s) created successfully\n", newDir.Name, newDir.ID)))
}

// delete a physical file on the server for the user
func (a *API) DeleteDir(w http.ResponseWriter, r *http.Request) {
	dir := r.Context().Value(Directory).(*svc.Directory)
	if err := a.Svc.RemoveDir(dir.DriveID, dir.ID); err != nil {
		http.Error(w, fmt.Sprintf("failed to remove directory: %v", err), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(fmt.Sprintf("\ndirectory %s (id=%s) deleted\n", dir.Name, dir.ID)))
}

// -------- drives --------------------------------

// sends drive metadata. does not return entire contents of drive.
func (a *API) GetDrive(w http.ResponseWriter, r *http.Request) {
	drive := r.Context().Value(Drive).(*svc.Drive)
	data, err := drive.ToJSON()
	if err != nil {
		http.Error(w, fmt.Sprintf("\nfailed to codify drive info to JSON: %v\n", err), http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// -------- sync ----------------------------------

// generate (or refresh) a sync index for a given drive
func (a *API) GenIndex(w http.ResponseWriter, r *http.Request) {
	driveID := r.Context().Value(Drive).(string)
	idx, err := a.Svc.GenSyncIndex(driveID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, err := idx.ToJSON()
	if err != nil {
		http.Error(w, fmt.Sprintf("\nfailed to encode sync index: %v\n", err), http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// retrieves (or generates) a sync index for a given drive.
// sync operations are coordinated on the client side, so the server
// only needs to manage indicies -- not coordinate operations.
func (a *API) GetIdx(w http.ResponseWriter, r *http.Request) {
	driveID := r.Context().Value(Drive).(string)
	idx, err := a.Svc.GetSyncIdx(driveID)
	if err != nil {
		http.Error(w, fmt.Sprintf("\nfailed to retrieve sync index: %v\n", err), http.StatusInternalServerError)
		return
	}
	if idx == nil { // no drive found
		http.Error(w, fmt.Sprintf("\ndrive %s not found\n", driveID), http.StatusNotFound)
		return
	}
	data, err := idx.ToJSON()
	if err != nil {
		http.Error(w, fmt.Sprintf("\nfailed to encode sync index: %v\n", err), http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// refreshes the server side Update map for this drives sync index.
// drives must already have been indexed prior to calling this endpoint.
func (a *API) GetUpdates(w http.ResponseWriter, r *http.Request) {
	driveID := r.Context().Value(Drive).(string)
	newIdx, err := a.Svc.RefreshUpdates(driveID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, err := newIdx.ToJSON()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
