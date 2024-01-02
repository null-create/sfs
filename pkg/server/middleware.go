package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/sfs/pkg/auth"
	svc "github.com/sfs/pkg/service"

	"github.com/go-chi/chi/v5"
)

// add json header to requests. added to middleware stack
// during router instantiation.
func ContentTypeJson(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json;charset=utf8")
		h.ServeHTTP(w, r)
	})
}

// -------- all item contexts ------------------------------------

func AllFilesCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")
		if userID == "" {
			http.Error(w, "no user ID provided", http.StatusBadRequest)
			return
		}
		files, err := getAllFiles(userID, getDBConn("Files"))
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to get files from database: %v", err), http.StatusInternalServerError)
			return
		}
		if len(files) == 0 {
			http.Error(w, fmt.Sprintf("no files found for user %s", userID), http.StatusNotFound)
			return
		}
		newCtx := context.WithValue(r.Context(), Files, files)
		h.ServeHTTP(w, r.WithContext(newCtx))
	})
}

// -------- new item/user context --------------------------------

// attempts to get filename, owner, and path from a requets
// context, then create a new file object to use for downloading
func NewFileCtx(h http.Handler) http.Handler {
	tokenValidator := auth.NewT()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// validate token
		fileToken := r.Header.Get("Authorization")
		if fileToken == "" {
			http.Error(w, "no authorization token provided", http.StatusBadRequest)
			return
		}
		fileInfo, err := tokenValidator.Verify(fileToken)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to verify file token: %v", err), http.StatusInternalServerError)
			return
		}
		// unmarshal new file data and check if it already exists before creating
		newFile, err := svc.UnmarshalFileStr(fileInfo)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to unmarshal file data: %v", err), http.StatusInternalServerError)
			return
		}
		file, err := findFile(newFile.ID, getDBConn("Files"))
		if err != nil {
			http.Error(w, "failed to query file database", http.StatusInternalServerError)
			return
		} else if file != nil {
			http.Error(w, fmt.Sprintf("file %s (id=%s) already exists", file.Name, file.ID), http.StatusBadRequest)
			return
		}
		newCtx := context.WithValue(r.Context(), File, newFile)
		h.ServeHTTP(w, r.WithContext(newCtx))
	})
}

func NewUserCtx(h http.Handler) http.Handler {
	tokenValidator := auth.NewT()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// validate request token
		userToken := r.Header.Get("Authorization")
		if userToken == "" {
			http.Error(w, "no authorization token provided", http.StatusBadRequest)
			return
		}
		userInfo, err := tokenValidator.Verify(userToken)
		if err != nil {
			msg := fmt.Sprintf("failed to verify user token: %v", err)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		// unmarshal data and check database before creating a new user
		newUser, err := auth.UnmarshalUser(userInfo)
		if err != nil {
			msg := fmt.Sprintf("failed to unmarshal user data: %v", err)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		// check if this user already exists before adding
		user, err := findUser(newUser.ID, getDBConn("Users"))
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to query database for user: %v", err), http.StatusInternalServerError)
			return
		} else if user != nil {
			http.Error(w, fmt.Sprintf("user %s (id=%s) already exists", newUser.Name, newUser.ID), http.StatusBadRequest)
			return
		}
		newCtx := context.WithValue(r.Context(), User, newUser)
		h.ServeHTTP(w, r.WithContext(newCtx))
	})
}

func NewDirectoryCtx(h http.Handler) http.Handler {
	tokenValidator := auth.NewT()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// validate token
		dirToken := r.Header.Get("Authorization")
		dirInfo, err := tokenValidator.Verify(dirToken)
		if err != nil {
			msg := fmt.Sprintf("failed to verify directory token: %v", err)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		newDir, err := svc.UnmarshalDirStr(dirInfo)
		if err != nil {
			msg := fmt.Sprintf("failed to unmarshal directory data: %v", err)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		// see if this directory is already in the DB first
		dir, err := findDir(newDir.ID, getDBConn("Directories"))
		if err != nil {
			http.Error(w, "failed to query directory database", http.StatusInternalServerError)
			return
		} else if dir != nil {
			http.Error(w, fmt.Sprintf("directory (id=%s) already exists", newDir.ID), http.StatusBadRequest)
			return
		}
		newCtx := context.WithValue(r.Context(), Directory, newDir)
		h.ServeHTTP(w, r.WithContext(newCtx))
	})
}

// ------- authentication --------------------------------

// retrieve jwt token from request & verify
func AuthenticateUser(reqToken string) (*auth.User, error) {
	tokenValidator := auth.NewT()
	userID, err := tokenValidator.Verify(reqToken)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %v", err)
	}
	// attempt to find data about the user from the the user db
	user, err := findUser(userID, getDBConn("Users"))
	if err != nil {
		return nil, fmt.Errorf("failed to query database for user: %v", err)
	} else if user == nil {
		return nil, fmt.Errorf("user (id=%s) not found", userID)
	}
	return user, nil
}

// get user info
func AuthUserHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqToken := r.Header.Get("Authorization")
		if reqToken == "" {
			http.Error(w, "header had no request token", http.StatusBadRequest)
			return
		}
		user, err := AuthenticateUser(reqToken)
		if err != nil {
			if strings.Contains(err.Error(), "failed to query database") {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			} else if strings.Contains(err.Error(), "user not found") {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		newCtx := context.WithValue(r.Context(), User, user)
		h.ServeHTTP(w, r.WithContext(newCtx))
	})
}

// ------ standard context --------------------------------

// single file context
func FileCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fileID := chi.URLParam(r, "fileID")
		if fileID == "" {
			http.Error(w, "fileID not set", http.StatusBadRequest)
			return
		}
		file, err := findFile(fileID, getDBConn("Files"))
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to retrieve file info: %v", err), http.StatusInternalServerError)
			return
		} else if file == nil {
			http.Error(w, fmt.Sprintf("file (id=%s) not found", fileID), http.StatusNotFound)
			return
		}
		if !file.Exists() {
			http.Error(w, "file was in database but physical file was not found", http.StatusInternalServerError)
			return
		}
		ctx := context.WithValue(r.Context(), File, file)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// multiple files context
func FilesCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")
		if userID == "" {
			http.Error(w, "userID is required", http.StatusBadRequest)
			return
		}
		// TODO: retrieve files list from request. need to figure out how to
		// send a list of files as part of a request.
	})
}

func DirCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dirID := chi.URLParam(r, "dirID")
		if dirID == "" {
			http.Error(w, "dirID not set", http.StatusBadRequest)
			return
		}
		dir, err := findDir(dirID, getDBConn("Directories"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if dir == nil {
			http.Error(w, fmt.Sprintf("directory (id=%s) not found", dirID), http.StatusNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), Directory, dir)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func DriveCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		driveID := chi.URLParam(r, "driveID")
		if driveID == "" {
			http.Error(w, "driveID not set", http.StatusBadRequest)
			return
		}
		drive, err := findDrive(driveID, getDBConn("Drives"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if drive == nil {
			http.Error(w, fmt.Sprintf("drive (id=%s) not found", driveID), http.StatusNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), Drive, drive)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// standard user context for established users.
// in conjunction with AuthUserHandler, which is part of the router's
// standard middleware stack.
func UserCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")
		if userID == "" {
			http.Error(w, "userID not set", http.StatusBadRequest)
			return
		}
		user, err := findUser(userID, getDBConn("Users"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if user == nil {
			http.Error(w, fmt.Sprintf("user (id=%s) not found", userID), http.StatusNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), User, user)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ------ admin stuff --------------------------------

func AdminOnly(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, ok := ctx.Value("acl.permission").(float64)
		if !ok {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	})
}
