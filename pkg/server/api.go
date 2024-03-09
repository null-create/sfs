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
	"github.com/sfs/pkg/logger"
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
	Svc       *Service       // SFS service instance
	log       *logger.Logger // API logging
}

// initialize sfs service
func NewAPI(newService bool, isAdmin bool) *API {
	svc, err := Init(newService, isAdmin)
	if err != nil {
		log.Fatalf("failed to initialize new service instance: %v", err)
	}
	return &API{
		StartTime: time.Now().UTC(),
		Svc:       svc,
		log:       logger.NewLogger("API"),
	}
}

// -------- testing ----------------------------------------------------

// placeholder handler for testing purposes
func (a *API) Placeholder(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("not implemented yet :("))
}

// --------- general ----------------------------------------------------

// generic response. sends msg with 200 and logs message.
func (a *API) write(w http.ResponseWriter, msg string) {
	a.log.Log("INFO", msg)
	w.Write([]byte(msg))
}

// not found response. sends a 404 and logs message.
func (a *API) notFoundError(w http.ResponseWriter, err string) {
	a.log.Warn(err)
	http.Error(w, err, http.StatusNotFound)
}

// sends a bad request (400) with error message, and logs message
func (a *API) clientError(w http.ResponseWriter, err string) {
	a.log.Warn(err)
	http.Error(w, err, http.StatusBadRequest)
}

// sends an internal server error (500) with an error message, and logs the message
func (a *API) serverError(w http.ResponseWriter, err string) {
	a.log.Error(err)
	http.Error(w, err, http.StatusInternalServerError)
}

// -------- users (admin only) -----------------------------------------

// add a new user and drive to sfs instance. user existance and
// struct pointer should be created by NewUser middleware
func (a *API) AddNewUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(User).(*auth.User)
	if err := a.Svc.AddUser(user); err != nil {
		if strings.Contains(err.Error(), "user") {
			a.clientError(w, err.Error()) // user already exists
			return
		} else {
			a.serverError(w, err.Error())
			return
		}
	} else {
		a.write(w, fmt.Sprintf("\nuser (name=%s id=%s) added\n", user.Name, user.ID))
	}
}

// send user metadata.
func (a *API) GetUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(User).(*auth.User)
	userData, err := user.ToJSON()
	if err != nil {
		a.serverError(w, err.Error())
		return
	}
	a.write(w, string(userData))
}

// return a list of all active users. used for admin and testing purposes.
func (a *API) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users := r.Context().Value(Users).([]*auth.User)
	for _, u := range users {
		data, err := u.ToJSON()
		if err != nil {
			a.serverError(w, err.Error())
			return
		}
		w.Write(data)
	}
}

// use UserCtx middleware before a call to this
func (a *API) UpdateUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(User).(*auth.User)
	if err := a.Svc.UpdateUser(user); err != nil {
		a.serverError(w, err.Error())
		return
	}
	a.write(w, "user updated")
}

// remove a user from the server
func (a *API) DeleteUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(User).(*auth.User)
	if err := a.Svc.RemoveUser(user.ID); err != nil {
		a.serverError(w, err.Error())
		return
	}
	a.write(w, fmt.Sprintf("user (name=%s id=%s) removed from server", user.Name, user.ID))
}

// -------- files -----------------------------------------

// get file metadata
func (a *API) GetFileInfo(w http.ResponseWriter, r *http.Request) {
	file := r.Context().Value(File).(*svc.File)
	data, err := file.ToJSON()
	if err != nil {
		a.serverError(w, fmt.Sprintf("failed to convert to JSON: %s", err.Error()))
		return
	}
	a.write(w, string(data))
}

// retrieve a file from the server
func (a *API) ServeFile(w http.ResponseWriter, r *http.Request) {
	// file existance was confirmed by the middleware at this point
	file := r.Context().Value(File).(*svc.File)

	// set the response header for the download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name))
	w.Header().Set("Content-Type", "application/octet-stream")

	// send the file
	http.ServeFile(w, r, file.ServerPath)
	a.log.Info(fmt.Sprintf("served file %s: %s", file.Name, file.ServerPath))
}

// get json blobs of all files available on the server for a user.
// only sends metadata, not the actual files.
func (a *API) GetAllFileInfo(w http.ResponseWriter, r *http.Request) {
	files := r.Context().Value(Files).([]*svc.File)
	for _, file := range files {
		data, err := file.ToJSON()
		if err != nil {
			a.serverError(w, fmt.Sprintf("failed to convert %s (id=%s) to JSON: %v", file.Name, file.ID, err))
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
		a.serverError(w, fmt.Sprintf("failed to copy file data into buffer: %v", err))
		return
	}
	defer r.Body.Close()
	// save file to server
	if err := newFile.Save(buf.Bytes()); err != nil {
		a.serverError(w, fmt.Sprintf("failed to save file to server: %v", err))
		return
	}
	// update service
	if err := a.Svc.AddFile(newFile.DirID, newFile); err != nil {
		a.log.Error(fmt.Sprintf("failed to add file to service: %v", err))
		// remove file from server.
		// don't want a file we don't have a record for.
		if err2 := os.Remove(newFile.ServerPath); err2 != nil {
			a.log.Error(fmt.Sprintf("failed to remove %s from server: %v\n", newFile.Name, err2))
		}
		return
	}
	a.write(w, fmt.Sprintf("%s has been added to the server", newFile.Name))
}

// update the file
func (a *API) putFile(w http.ResponseWriter, r *http.Request, file *svc.File) {
	// retrieve file data from request body
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r.Body)
	if err != nil {
		a.serverError(w, fmt.Sprintf("failed to download form file data: %v", err))
		return
	}
	// update file
	if err := a.Svc.UpdateFile(file, buf.Bytes()); err != nil {
		a.serverError(w, fmt.Sprintf("failed to update %s (id=%s)", file.Name, file.ID))
		return
	}
	a.write(w, fmt.Sprintf("%s updated (owner id=%s)", file.Name, file.OwnerID))
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
		a.serverError(w, fmt.Sprintf("\nfailed to delete file: %v\n", err))
		return
	}
	a.write(w, fmt.Sprintf("%s (id=%s) deleted from server", file.Name, file.ID))
}

// ------- directories --------------------------------

// temp for testing
func (a *API) GetAllDirsInfo(w http.ResponseWriter, r *http.Request) {
	dirs := r.Context().Value(Directories).([]*svc.Directory)
	for _, dir := range dirs {
		data, err := dir.ToJSON()
		if err != nil {
			a.serverError(w, err.Error())
			return
		}
		w.Write(data)
	}
}

// returns metadata for a single directory (not its children).
func (a *API) GetDirInfo(w http.ResponseWriter, r *http.Request) {
	dir := r.Context().Value(Directory).(*svc.Directory)
	data, err := dir.ToJSON()
	if err != nil {
		a.serverError(w, err.Error())
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
			a.log.Error(err.Error())
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
	archive := filepath.Join(dir.Path, dir.Name+".zip")
	if err := transfer.Zip(dir.Path, archive); err != nil {
		a.serverError(w, fmt.Sprintf("failed to compress directory: %v", err))
		return
	}
	// send archive file
	http.ServeFile(w, r, archive)
	a.log.Info(fmt.Sprintf("served file: %v", archive))
	// remove tmp archive file
	if err := os.Remove(archive); err != nil {
		a.log.Error(fmt.Sprintf("failed to remove temp archive %s: %v", archive, err))
	}
}

// update the directory on the server
func (a *API) PutDir(w http.ResponseWriter, r *http.Request) {
	dir := r.Context().Value(Directory).(*svc.Directory)
	if err := a.Svc.UpdateDir(dir.DriveID, dir); err != nil {
		a.serverError(w, err.Error())
		return
	}
	a.write(w, fmt.Sprintf("directory (id=%s) has been updated", dir.ID))
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
		a.serverError(w, fmt.Sprintf("failed to create directory: %v", err))
		return
	}
	a.write(w, fmt.Sprintf("directory %s (id=%s) created successfully", newDir.Name, newDir.ID))
}

// delete a physical file on the server for the user
func (a *API) DeleteDir(w http.ResponseWriter, r *http.Request) {
	dir := r.Context().Value(Directory).(*svc.Directory)
	if err := a.Svc.RemoveDir(dir.DriveID, dir.ID); err != nil {
		a.serverError(w, fmt.Sprintf("failed to remove directory: %v", err))
		return
	}
	a.write(w, fmt.Sprintf("directory %s (id=%s) deleted", dir.Name, dir.ID))
}

// -------- drives --------------------------------

/*
NOTE: drive create/delete are handled by user API functions.
*/

// sends drive metadata. does not return entire contents of drive.
func (a *API) GetDrive(w http.ResponseWriter, r *http.Request) {
	drive := r.Context().Value(Drive).(*svc.Drive)
	data, err := drive.ToJSON()
	if err != nil {
		a.serverError(w, fmt.Sprintf("failed to codify drive info to JSON: %v", err))
		return
	}
	w.Write(data)
}

// add a new drive to the server. used as part of a separate registration process.
func (a *API) NewDrive(w http.ResponseWriter, r *http.Request) {
	drive := r.Context().Value(Drive).(*svc.Drive)
	if err := a.Svc.AddDrive(drive); err != nil {
		// this shouldn't happen but just in case
		if strings.Contains(err.Error(), "already registered") {
			a.write(w, err.Error())
			return
		}
		a.serverError(w, err.Error())
		return
	}
	a.write(w, fmt.Sprintf("drive (id=%s) added successfully", drive.ID))
}

// -------- sync ----------------------------------

// generate (or refresh) a sync index for a given drive
func (a *API) GenIndex(w http.ResponseWriter, r *http.Request) {
	driveID := r.Context().Value(Drive).(string)
	idx, err := a.Svc.GenSyncIndex(driveID)
	if err != nil {
		a.serverError(w, err.Error())
		return
	}
	data, err := idx.ToJSON()
	if err != nil {
		a.serverError(w, fmt.Sprintf("failed to encode sync index: %v", err))
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
		a.serverError(w, fmt.Sprintf("failed to retrieve sync index: %v", err))
		return
	}
	if idx == nil { // no drive found
		a.notFoundError(w, fmt.Sprintf("drive %s not found", driveID))
		return
	}
	data, err := idx.ToJSON()
	if err != nil {
		a.serverError(w, fmt.Sprintf("failed to encode sync index: %v", err))
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
		a.serverError(w, err.Error())
		return
	}
	data, err := newIdx.ToJSON()
	if err != nil {
		a.serverError(w, err.Error())
		return
	}
	w.Write(data)
}
