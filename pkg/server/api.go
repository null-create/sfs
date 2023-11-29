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

	"github.com/go-chi/chi"
)

// TODO: add general logging

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

/*
NOTE: r *http.Request contains a context field, which
can be populated using context middleware. if context
middleware is successful, then we can retrieve values
like this:

ctx := r.Context()

see: https://pkg.go.dev/net/http#Request.Context

*/

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
// handles errors and not found cases. returns nil if either of these
// are the case, otherwise returns a file pointer.
func (a *API) findF(w http.ResponseWriter, r *http.Request) *svc.File {
	fileID := chi.URLParam(r, "fileID")
	f, err := a.Svc.Db.GetFile(fileID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	} else if f == nil {
		return nil
	}
	return f
}

// get file metadata
func (a *API) GetFileInfo(w http.ResponseWriter, r *http.Request) {
	f := a.findF(w, r)
	if f == nil {
		fileID := chi.URLParam(r, "fileID")
		http.Error(w, fmt.Sprintf("file (id=%s) not found", fileID), http.StatusNotFound)
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
		fileID := chi.URLParam(r, "fileID")
		msg := fmt.Sprintf("file (id=%s) not found", fileID)
		http.Error(w, msg, http.StatusNotFound)
		return
	}

	// Set the response header for the download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", f.Name))
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, f.ServerPath)
}

// get json blobs of all files available on the server
func (a *API) GetAllFiles(w http.ResponseWriter, r *http.Request) {
	// TODO: replace with user-specific get-files db call
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

func (a *API) newFile(w http.ResponseWriter, r *http.Request) {
	formFile, _, err := r.FormFile("myFile")
	if err != nil {
		msg := fmt.Sprintf("failed to retrive form file data: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	defer formFile.Close()

	// retrieve new file object from context before copying data
	ctx := r.Context()
	newFile := ctx.Value(File).(*svc.File)

	// retrieve file
	var buf bytes.Buffer
	_, err = io.Copy(&buf, formFile)
	if err != nil {
		msg := fmt.Sprintf("failed to retrive form file data: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	if err := newFile.Save(buf.Bytes()); err != nil {
		msg := fmt.Sprintf("failed to download file to server: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	// update checksum
	newFile.LastSync = time.Now().UTC()
	newFile.CheckSum, err = svc.CalculateChecksum(newFile.ServerPath, "sha256")
	if err != nil {
		msg := fmt.Sprintf("failed to calculate checksum %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	a.Svc.Db.WhichDB("files")
	if err := a.Svc.Db.AddFile(newFile); err != nil {
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

	// update file
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
	// we ignore the error since we don't want to return it to the user
	if err = f.UpdateChecksum(); err != nil {
		log.Printf("failed to update checksum: %v", err)
	}

	// update DB
	if err := a.Svc.Db.UpdateFile(f); err != nil {
		msg := fmt.Sprintf("failed to update file database: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
}

// upload or update a file on/to the server
func (a *API) PutFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut { // update the file
		ctx := r.Context()
		f := ctx.Value(File).(*svc.File)
		a.putFile(w, r, f)
	} else if r.Method == http.MethodPost { // create a new file.
		a.newFile(w, r)
	}
}

// delete a file from the server
func (a *API) DeleteFile(w http.ResponseWriter, r *http.Request) {
	f := a.findF(w, r)
	if f == nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	// remove physical file
	// if err := os.Remove(f.ServerPath); err != nil {
	// 	msg := fmt.Sprintf("failed to remove file from server: %v", err)
	// 	http.Error(w, msg, http.StatusInternalServerError)
	// 	return
	// }

	// TODO: use/create a.Svc.DeleteFile() instead of os.Remove()
	// using the os package directly is way too risky.

	// remove from database
	if err := a.Svc.Db.RemoveFile(f.ID); err != nil {
		msg := fmt.Sprintf("failed to remove file from database: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	log.Printf("[INFO] filed (%s) deleted from server", f.Path)
	w.Write([]byte(fmt.Sprintf("file (%s) deleted", f.Name)))
}

// ------- directories --------------------------------

// check the db for the existence of a directory.
//
// handles errors and not found cases. returns nil if either of these
// are the case, otherwise returns a directory pointer.
func (a *API) findD(w http.ResponseWriter, r *http.Request) *svc.Directory {
	dirID := chi.URLParam(r, "dirID")
	a.Svc.Db.WhichDB("Directories")
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
	dirName := ctx.Value(Name).(string)
	parentID := ctx.Value(Parent).(string)
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
	dir := svc.NewDirectory(dirName, owner, path)

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

// -------- sync ----------------------------------
