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
		log:       logger.NewLogger("API", "None"),
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
	a.log.Log(logger.INFO, msg)
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

// returns a user struct for a new or existing user, assuming it exists in the server database.
func (a *API) getNewUserFromRequest(r *http.Request) (*auth.User, error) {
	user := r.Context().Value(User).(*auth.User)
	if user == nil {
		return nil, fmt.Errorf("no user in request")
	}
	return user, nil
}

// get a user from the ID supplied by the request. returns nil if not found.
func (a *API) getUserFromRequest(r *http.Request) (*auth.User, error) {
	userID := r.Context().Value(User).(string)
	if userID == "" {
		return nil, fmt.Errorf("no user ID string found")
	}
	user, err := a.Svc.GetUser(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user (id=%s) not found", userID)
	}
	return user, nil
}

// add a new user and drive to sfs instance. user existance and
// struct pointer should be created by NewUser middleware
func (a *API) AddNewUser(w http.ResponseWriter, r *http.Request) {
	user, err := a.getNewUserFromRequest(r)
	if err != nil {
		a.serverError(w, err.Error())
		return
	}
	if err := a.Svc.AddUser(user); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			a.write(w, "user is already registered") // user already exists
			return
		} else {
			a.serverError(w, err.Error())
			return
		}
	} else {
		a.write(w, fmt.Sprintf("user (name=%s id=%s) added", user.Name, user.ID))
	}
}

// send user metadata.
func (a *API) GetUser(w http.ResponseWriter, r *http.Request) {
	user, err := a.getUserFromRequest(r)
	if err != nil {
		if strings.Contains(err.Error(), "user") { // non-existing user or missing ID errors
			a.clientError(w, err.Error())
		} else {
			a.serverError(w, err.Error())
		}
		return
	}
	userData, err := user.ToJSON()
	if err != nil {
		a.serverError(w, err.Error())
		return
	}
	a.write(w, string(userData))
}

// return a list of all active users. used for admin and testing purposes.
func (a *API) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := a.Svc.Db.GetUsers()
	if err != nil {
		a.serverError(w, err.Error())
		return
	}
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
	user, err := a.getUserFromRequest(r)
	if err != nil {
		if strings.Contains(err.Error(), "user") { // non-existing user or missing ID errors
			a.clientError(w, err.Error())
		} else {
			a.serverError(w, err.Error())
		}
		return
	}
	if err := a.Svc.UpdateUser(user); err != nil {
		a.serverError(w, err.Error())
		return
	}
	a.write(w, fmt.Sprintf("user (name=%s id=%s) updated", user.Name, user.ID))
}

// remove a user from the server
func (a *API) DeleteUser(w http.ResponseWriter, r *http.Request) {
	user, err := a.getUserFromRequest(r)
	if err != nil {
		if strings.Contains(err.Error(), "user") { // non-existing user or missing ID errors
			a.clientError(w, err.Error())
		} else {
			a.serverError(w, err.Error())
		}
		return
	}
	if err := a.Svc.RemoveUser(user.ID); err != nil {
		a.serverError(w, err.Error())
		return
	}
	a.write(w, fmt.Sprintf("user (name=%s id=%s) removed from server", user.Name, user.ID))
}

// -------- files -----------------------------------------

func (a *API) getNewFileFromRequest(r *http.Request) (*svc.File, error) {
	file := r.Context().Value(File).(*svc.File)
	if file == nil {
		return nil, fmt.Errorf("file object not found in request")
	}
	return file, nil
}

// gets a file from the ID provided by the request. returns nil if not found.
func (a *API) getFileFromRequest(r *http.Request) (*svc.File, error) {
	fileID := r.Context().Value(File).(string)
	if fileID == "" {
		return nil, fmt.Errorf("no file ID found in request")
	}
	file, err := a.Svc.Db.GetFileByID(fileID)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, fmt.Errorf("file (id=%s) not found", fileID)
	}
	return file, nil
}

// get file metadata
func (a *API) GetFileInfo(w http.ResponseWriter, r *http.Request) {
	f, err := a.getFileFromRequest(r)
	if err != nil {
		if strings.Contains(err.Error(), "file") { // not found or missing ID errors
			a.clientError(w, err.Error())
		} else {
			a.serverError(w, err.Error())
		}
		return
	}
	file, err := a.Svc.GetFile(f.DriveID, f.ID)
	if err != nil {
		a.serverError(w, err.Error())
		return
	}
	data, err := file.ToJSON()
	if err != nil {
		a.serverError(w, "failed to convert to JSON: "+err.Error())
		return
	}
	a.write(w, string(data))
}

// retrieve a file from the server
func (a *API) ServeFile(w http.ResponseWriter, r *http.Request) {
	f, err := a.getFileFromRequest(r)
	if err != nil {
		if strings.Contains(err.Error(), "file") { // not found or missing ID errors
			a.clientError(w, err.Error())
		} else {
			a.serverError(w, err.Error())
		}
		return
	}
	file, err := a.Svc.GetFile(f.DriveID, f.ID)
	if err != nil {
		a.serverError(w, err.Error())
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name))
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, file.ServerPath)
	a.log.Info(fmt.Sprintf("served file %s: %s", file.Name, file.ServerPath))
}

// get json blobs of all files available on the server for a user.
// only sends metadata, not the actual files.
func (a *API) GetAllFileInfo(w http.ResponseWriter, r *http.Request) {
	user, err := a.getUserFromRequest(r)
	if err != nil {
		if strings.Contains(err.Error(), "user") {
			a.clientError(w, err.Error())
		} else {
			a.serverError(w, err.Error())
		}
		return
	}
	files, err := a.Svc.GetAllFiles(user.DriveID)
	if err != nil {
		a.serverError(w, err.Error())
		return
	}
	for _, file := range files {
		data, err := file.ToJSON()
		if err != nil {
			a.serverError(w, fmt.Sprintf("failed to convert %s (id=%s) to JSON: %v", file.Name, file.ID, err))
			return
		}
		w.Write(data)
	}
}

// add initial file metadata to the server. creates an empty files,
// does not create file contents, though svc.AddFile() does attempt
// to write out the data. This will be remidied in a future version.
func (a *API) newFile(w http.ResponseWriter, r *http.Request, newFile *svc.File) {
	f, err := a.Svc.Db.GetFileByID(newFile.ID)
	if err != nil {
		a.serverError(w, err.Error())
		return
	}
	if f != nil {
		a.clientError(w, fmt.Sprintf("file (name=%s id=%s) already exists", newFile.Name, newFile.ID))
		return
	}
	if err := a.Svc.AddFile(newFile.DirID, newFile); err != nil {
		a.serverError(w, fmt.Sprintf("failed to add %s to service: %v", newFile.Name, err))
		return
	}
	a.write(w, fmt.Sprintf("file (%s) has been added to the server", newFile.Name))
}

// update the file on the server
func (a *API) putFile(w http.ResponseWriter, r *http.Request, file *svc.File) {
	f, _, err := r.FormFile("myFile")
	if err != nil {
		a.serverError(w, "failed to retrieve form file: "+err.Error())
		return
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, f)
	if err != nil {
		a.serverError(w, "failed to copy file: "+err.Error())
		return
	}
	if err := f.Close(); err != nil {
		a.serverError(w, "failed to close form file: "+err.Error())
		return
	}

	if err := a.Svc.UpdateFile(file, buf.Bytes()); err != nil {
		a.serverError(w, fmt.Sprintf("failed to update %s (id=%s): %v", file.Name, file.ID, err))
		return
	}
	a.write(w, fmt.Sprintf("file (%s) updated (owner id=%s)", file.Name, file.OwnerID))
}

// upload or update a file on/to the server
func (a *API) PutFile(w http.ResponseWriter, r *http.Request) {
	file, err := a.getNewFileFromRequest(r)
	if err != nil {
		a.serverError(w, err.Error())
		return
	}
	if r.Method == http.MethodPut { // update the file
		a.putFile(w, r, file)
	} else if r.Method == http.MethodPost { // create a new file.
		a.newFile(w, r, file)
	}
}

// delete a file from the server
func (a *API) DeleteFile(w http.ResponseWriter, r *http.Request) {
	file, err := a.getNewFileFromRequest(r)
	if err != nil {
		a.serverError(w, err.Error())
		return
	}
	if err := a.Svc.DeleteFile(file); err != nil {
		a.serverError(w, "failed to delete file: "+err.Error())
		return
	}
	a.write(w, fmt.Sprintf("%s (id=%s) deleted from server", file.Name, file.ID))
}

// ------- directories --------------------------------

// used by functions that are creating new objects. requests will
// send entire objects from the client when new items need to be registered.
func (a *API) getNewDirFromRequest(r *http.Request) (*svc.Directory, error) {
	dir := r.Context().Value(Directory).(*svc.Directory)
	if dir == nil {
		return nil, fmt.Errorf("dir not found")
	}
	return dir, nil
}

// used by functions working with registered directories
func (a *API) getDirFromRequest(r *http.Request) (*svc.Directory, error) {
	dirID := r.Context().Value(Directory).(string)
	if dirID == "" {
		return nil, fmt.Errorf("no directory specified")
	}
	dir, err := a.Svc.GetDirByID(dirID)
	if err != nil {
		return nil, err
	}
	if dir == nil {
		return nil, fmt.Errorf("directory (id=%s) not found", dirID)
	}
	return dir, nil
}

// temp for testing
func (a *API) GetAllDirsInfo(w http.ResponseWriter, r *http.Request) {
	dirs, err := a.Svc.Db.GetAllDirectories()
	if err != nil {
		a.serverError(w, err.Error())
		return
	}
	for _, dir := range dirs {
		data, err := dir.ToJSON()
		if err != nil {
			a.serverError(w, err.Error())
			return
		}
		w.Write(data)
	}
}

func (a *API) GetUsersDirs(w http.ResponseWriter, r *http.Request) {
	user, err := a.getNewUserFromRequest(r)
	if err != nil {
		a.serverError(w, err.Error())
		return
	}
	dirs, err := a.Svc.GetAllDirs(user.DriveID)
	if err != nil {
		a.serverError(w, err.Error())
		return
	}
	if dirs == nil {
		a.notFoundError(w, fmt.Sprintf("no directories found for user (name=%s id=%s)", user.Name, user.ID))
		return
	}
	for _, dir := range dirs {
		data, err := dir.ToJSON()
		if err != nil {
			a.serverError(w, err.Error())
			return
		}
		w.Write(data)
	}
}

func (a *API) walkDir(w http.ResponseWriter, dir *svc.Directory) error {
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
	dir, err := a.getDirFromRequest(r)
	if err != nil {
		if strings.Contains(err.Error(), "directory") {
			a.clientError(w, err.Error()) // no directory or missing ID errors
		} else {
			a.serverError(w, err.Error())
		}
		return
	}
	// populate metadata
	dir = a.Svc.Populate(dir)
	// walk the directory tree starting from this directory, and
	// send JSON blobs for each object it discovers along the way.
	// data is sent in depth-first search order so client side
	// will need to sort that out.
	a.walkDir(w, dir)
}

// returns metadata for a single directory (not its children).
func (a *API) GetDirInfo(w http.ResponseWriter, r *http.Request) {
	dir, err := a.getDirFromRequest(r)
	if err != nil {
		if strings.Contains(err.Error(), "directory") {
			a.clientError(w, err.Error()) // no directory or missing ID errors
		} else {
			a.serverError(w, err.Error())
		}
		return
	}
	data, err := dir.ToJSON()
	if err != nil {
		a.serverError(w, err.Error())
		return
	}
	w.Write(data)
}

// retrieve a zipfile of the directory (and all its children)
func (a *API) GetDir(w http.ResponseWriter, r *http.Request) {
	dir, err := a.getDirFromRequest(r)
	if err != nil {
		if strings.Contains(err.Error(), "directory") {
			a.clientError(w, err.Error()) // no directory or missing ID errors
		} else {
			a.serverError(w, err.Error())
		}
		return
	}

	// create a tmp .zip file so we can transfer the directory and its contents
	archive := filepath.Join(dir.ServerPath, dir.Name+".zip")
	if err := transfer.Zip(dir.ServerPath, archive); err != nil {
		a.serverError(w, fmt.Sprintf("failed to compress directory: %v", err))
		return
	}

	// send archive file
	http.ServeFile(w, r, archive)

	// remove tmp archive file
	if err := os.Remove(archive); err != nil {
		a.log.Error(fmt.Sprintf("failed to remove temp archive %s: %v", archive, err))
	}
}

// update the directory on the server
func (a *API) PutDir(w http.ResponseWriter, r *http.Request) {
	dir, err := a.getDirFromRequest(r)
	if err != nil {
		if strings.Contains(err.Error(), "directory") {
			a.clientError(w, err.Error()) // no directory or missing ID errors
		} else {
			a.serverError(w, err.Error())
		}
		return
	}
	if err := a.Svc.UpdateDir(dir.DriveID, dir); err != nil {
		a.serverError(w, err.Error())
		return
	}
	a.write(w, fmt.Sprintf("directory (id=%s) has been updated", dir.ID))
}

// create a new empty physical directory on the server for a user
func (a *API) NewDir(w http.ResponseWriter, r *http.Request) {
	newDir, err := a.getNewDirFromRequest(r)
	if err != nil {
		a.serverError(w, err.Error())
		return
	}
	d, err := a.Svc.Db.GetDirectoryByID(newDir.ID)
	if err != nil {
		a.serverError(w, err.Error())
		return
	}
	if d != nil {
		a.clientError(w, fmt.Sprintf("directory (name=%s id=%s) already registered", newDir.Name, newDir.ID))
		return
	}
	if err := a.Svc.NewDir(newDir.DriveID, newDir.ParentID, newDir); err != nil {
		a.serverError(w, fmt.Sprintf("failed to create directory: %v", err))
		return
	}
	a.write(w, fmt.Sprintf("directory %s (id=%s) created successfully", newDir.Name, newDir.ID))
}

// delete a physical file on the server for the user
func (a *API) DeleteDir(w http.ResponseWriter, r *http.Request) {
	dir, err := a.getDirFromRequest(r)
	if err != nil {
		if strings.Contains(err.Error(), "directory") {
			a.clientError(w, err.Error()) // no directory or missing ID errors
		} else {
			a.serverError(w, err.Error())
		}
		return
	}
	if err := a.Svc.RemoveDir(dir.DriveID, dir.ID); err != nil {
		a.serverError(w, fmt.Sprintf("failed to remove directory: %v", err))
		return
	}
	a.write(w, fmt.Sprintf("directory %s (id=%s) deleted", dir.Name, dir.ID))
}

// -------- drives --------------------------------

func (a *API) getDriveIDFromRequest(r *http.Request) (string, error) {
	driveID := r.Context().Value(Drive).(string)
	if driveID == "" {
		return "", fmt.Errorf("no drive ID found in request")
	}
	return driveID, nil
}

func (a *API) getDriveFromRequest(r *http.Request) (*svc.Drive, error) {
	drive := r.Context().Value(Drive).(*svc.Drive)
	if drive == nil {
		return nil, fmt.Errorf("no drive found in request")
	}
	return drive, nil
}

// sends drive metadata. does not return entire contents of drive.
func (a *API) GetDrive(w http.ResponseWriter, r *http.Request) {
	driveID, err := a.getDriveIDFromRequest(r)
	if err != nil {
		a.clientError(w, err.Error())
		return
	}
	drive, err := a.Svc.LoadDrive(driveID)
	if err != nil {
		a.serverError(w, err.Error())
		return
	}
	data, err := drive.ToJSON()
	if err != nil {
		a.serverError(w, fmt.Sprintf("failed to codify drive info to JSON: %v", err))
		return
	}
	w.Write(data)
}

// add a new drive to the server. used as part of a separate registration process.
func (a *API) NewDrive(w http.ResponseWriter, r *http.Request) {
	newDrive, err := a.getDriveFromRequest(r)
	if err != nil {
		a.serverError(w, err.Error())
		return
	}
	if err := a.Svc.AddDrive(newDrive); err != nil {
		if strings.Contains(err.Error(), "already registered") {
			a.write(w, "drive is already registered")
			return
		}
		a.serverError(w, err.Error())
		return
	}
	a.write(w, fmt.Sprintf("drive (id=%s) added successfully", newDrive.ID))
}

// -------- sync ----------------------------------

// generate (or refresh) a sync index for a given drive.
// returns the index to the caller.
func (a *API) GenIndex(w http.ResponseWriter, r *http.Request) {
	drv, err := a.getDriveFromRequest(r)
	if err != nil {
		a.serverError(w, err.Error())
		return
	}
	if drv == nil {
		a.clientError(w, fmt.Sprintf("drive (id=%s) not found", drv.ID))
		return
	}
	newIdx, err := a.Svc.GenSyncIndex(drv.ID)
	if err != nil {
		a.serverError(w, err.Error())
		return
	}
	data, err := newIdx.ToJSON()
	if err != nil {
		a.serverError(w, fmt.Sprintf("failed to encode sync index: %v", err))
		return
	}
	w.Write(data)
	drv.SyncIndex = newIdx
	if err := a.Svc.UpdateDrive(drv); err != nil {
		a.serverError(w, fmt.Sprintf("failed to update drive (id=%s): %v", drv.ID, err))
		return
	}
}

// retrieves a sync index for a given drive.
// sync operations are coordinated on the client side, so the server
// only needs to manage indicies -- not coordinate operations.
func (a *API) GetIdx(w http.ResponseWriter, r *http.Request) {
	drv, err := a.getDriveFromRequest(r)
	if err != nil {
		a.serverError(w, err.Error())
		return
	}
	if drv == nil {
		a.clientError(w, fmt.Sprintf("drive (id=%s) not found", drv.ID))
		return
	}
	if drv.SyncIndex == nil {
		_, err := a.Svc.LoadDrive(drv.ID)
		if err != nil {
			a.serverError(w, err.Error())
			return
		}
	}
	data, err := drv.SyncIndex.ToJSON()
	if err != nil {
		a.serverError(w, fmt.Sprintf("failed to encode sync index: %v", err))
		return
	}
	w.Write(data)
}

// refreshes the server side Update map for this drives sync index.
// drives must already have been indexed prior to calling this endpoint.
func (a *API) GetUpdates(w http.ResponseWriter, r *http.Request) {
	drv, err := a.getDriveFromRequest(r)
	if err != nil {
		a.serverError(w, err.Error())
		return
	}
	if drv == nil {
		a.clientError(w, fmt.Sprintf("drive (id=%s) not found", drv.ID))
		return
	}
	newIdx, err := a.Svc.RefreshUpdates(drv.ID)
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
	drv.SyncIndex = newIdx
	if err := a.Svc.UpdateDrive(drv); err != nil {
		a.serverError(w, fmt.Sprintf("failed to update drive (id=%s): %v", drv.ID, err))
		return
	}
}
