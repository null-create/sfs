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

func AllUsersFilesCtx(h http.Handler) http.Handler {
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
			w.Write([]byte(fmt.Sprintf("no files found for user (id=%s)", userID)))
			return
		}
		newCtx := context.WithValue(r.Context(), Files, files)
		h.ServeHTTP(w, r.WithContext(newCtx))
	})
}

func AllFilesCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		files, err := getAllTheFiles(getDBConn("Files"))
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to get files from database: %v", err), http.StatusInternalServerError)
			return
		}
		if len(files) == 0 {
			w.Write([]byte("no files found"))
			return
		}
		newCtx := context.WithValue(r.Context(), Files, files)
		h.ServeHTTP(w, r.WithContext(newCtx))
	})
}

func AllUsersCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")
		if userID == "" {
			http.Error(w, "no user ID provided", http.StatusBadRequest)
			return
		}
		users, err := getAllUsers(userID, getDBConn("Users"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(users) == 0 {
			w.Write([]byte("no users found"))
			return
		}
		newCtx := context.WithValue(r.Context(), Users, users)
		h.ServeHTTP(w, r.WithContext(newCtx))
	})
}

func AllDirsCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dirs, err := findAllTheDirs(getDBConn("Directories"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(dirs) == 0 {
			w.Write([]byte("no directories found"))
			return
		}
		newCtx := context.WithValue(r.Context(), Directories, dirs)
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
		rawToken := r.Header.Get("Authorization")
		if rawToken == "" {
			http.Error(w, "no authorization token provided", http.StatusBadRequest)
			return
		}
		fileToken, err := tokenValidator.Extract(rawToken)
		if err != nil {
			http.Error(w, fmt.Sprintf("authorization failed: %v", err), http.StatusBadRequest)
			return
		}
		fileInfo, err := tokenValidator.Verify(fileToken)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to verify token: %v", err), http.StatusInternalServerError)
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
		rawToken := r.Header.Get("Authorization")
		if rawToken == "" {
			http.Error(w, "no authorization token provided", http.StatusBadRequest)
			return
		}
		userToken, err := tokenValidator.Extract(rawToken)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to extract token: %v", err), http.StatusBadRequest)
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// check if this user already exists before adding
		user, err := findUser(newUser.ID, getDBConn("Users"))
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to query database for user: %v", err), http.StatusInternalServerError)
			return
		} else if user != nil {
			w.Write([]byte(fmt.Sprintf("%s (id=%s) already exists", newUser.Name, newUser.ID)))
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
		rawToken := r.Header.Get("Authorization")
		if rawToken == "" {
			http.Error(w, "missing directory token", http.StatusBadRequest)
			return
		}
		dirToken, err := tokenValidator.Extract(rawToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		dirInfo, err := tokenValidator.Verify(dirToken)
		if err != nil {
			msg := fmt.Sprintf("failed to verify directory token: %v", err)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		// create new directory object
		newDir, err := svc.UnmarshalDirStr(dirInfo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

func NewDriveCtx(h http.Handler) http.Handler {
	tokenValidator := auth.NewT()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// validate token
		rawToken := r.Header.Get("Authorization")
		if rawToken == "" {
			http.Error(w, "missing drive token", http.StatusBadRequest)
			return
		}
		drvToken, err := tokenValidator.Extract(rawToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		drvInfo, err := tokenValidator.Verify(drvToken)
		if err != nil {
			msg := fmt.Sprintf("failed to verify new drive token: %v", err)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		// unmarshal into new drive object, and check whether its already registered
		newDrive, err := svc.UnmarshalDriveString(drvInfo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// see if this drive is already in the DB first
		drv, err := findDrive(newDrive.ID, getDBConn("Drives"))
		if err != nil {
			http.Error(w, "failed to query drive database", http.StatusInternalServerError)
			return
		} else if drv != nil {
			// we return 200 becaues if a drive already exists then it is registered,
			// and the client typically checks if its been registered on start up
			w.Write([]byte(fmt.Sprintf("drive (id=%s) already exists", newDrive.ID)))
			return
		}

		// serve
		newCtx := context.WithValue(r.Context(), Drive, newDrive)
		h.ServeHTTP(w, r.WithContext(newCtx))
	})
}

// ------- authentication --------------------------------

// retrieve jwt token from request & verify
func AuthenticateUser(reqToken string) (*auth.User, error) {
	tokenValidator := auth.NewT()
	userID, err := tokenValidator.Verify(reqToken)
	if err != nil {
		return nil, fmt.Errorf("failed to validate user token: %v", err)
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

func DriveIdCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		driveID := chi.URLParam(r, "driveID")
		if driveID == "" {
			http.Error(w, "driveID not set", http.StatusBadRequest)
			return
		}
		// verify the drive exists
		drive, err := findDrive(driveID, getDBConn("Drives"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if drive == nil {
			http.Error(w, fmt.Sprintf("drive (id=%s) not found", driveID), http.StatusNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), Drive, driveID)
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
