package server

import (
	"context"
	"fmt"
	"log"
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

func ContentTypeText(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		h.ServeHTTP(w, r)
	})
}

func ContentTypeCSS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "text/css")
		h.ServeHTTP(w, r)
	})
}

func EnableCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		h.ServeHTTP(w, r)
	})
}

/*
TODO:
Rethink what the middleware should actually do. There are several functions
that access the database and execute operations that could be handled directly
by the SFS service.

Middleware should probably just validate the JWT's coming in from clients,
and not mucn else. Let the service deal with service objects like files and directories.
*/

// -------- all item contexts ------------------------------------

func AllUsersFilesCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")
		if userID == "" {
			http.Error(w, "no user ID provided", http.StatusBadRequest)
			return
		}
		newCtx := context.WithValue(r.Context(), User, userID)
		h.ServeHTTP(w, r.WithContext(newCtx))
	})
}

// -------- new item/user context --------------------------------

func ClientNewFileCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fileID := chi.URLParam(r, "filePath")
		if fileID == "" {
			http.Error(w, "filePath not set", http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(r.Context(), File, fileID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func DiscoverCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		folderPath := chi.URLParam(r, "folderPath")
		if folderPath == "" {
			http.Error(w, "filePath not set", http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(r.Context(), Directory, folderPath)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// attempts to get filename, owner, and path from a requets
// context, then create a new file object to use for downloading
func NewFileCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenValidator := auth.NewT()
		fileInfo, err := tokenValidator.Validate(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to verify token: %v", err), http.StatusInternalServerError)
			return
		}
		newFile, err := svc.UnmarshalFileStr(fileInfo)
		if err != nil {
			if err.Error() == "invalid token" {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, fmt.Sprintf("failed to unmarshal file data: %v", err), http.StatusInternalServerError)
			return
		}
		newCtx := context.WithValue(r.Context(), File, newFile)
		h.ServeHTTP(w, r.WithContext(newCtx))
	})
}

func NewUserCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenValidator := auth.NewT()
		userInfo, err := tokenValidator.Validate(r)
		if err != nil {
			if err.Error() == "invalid token" {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			msg := fmt.Sprintf("failed to verify user token: %v", err)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		newUser, err := auth.UnmarshalUserStr(userInfo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		newCtx := context.WithValue(r.Context(), User, newUser)
		h.ServeHTTP(w, r.WithContext(newCtx))
	})
}

func NewDirectoryCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenValidator := auth.NewT()
		dirInfo, err := tokenValidator.Validate(r)
		if err != nil {
			if err.Error() == "invalid token" {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			msg := fmt.Sprintf("failed to verify directory token: %v", err)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		newDir, err := svc.UnmarshalDirStr(dirInfo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		newCtx := context.WithValue(r.Context(), Directory, newDir)
		h.ServeHTTP(w, r.WithContext(newCtx))
	})
}

func NewDriveCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenValidator := auth.NewT()
		drvInfo, err := tokenValidator.Validate(r)
		if err != nil {
			if err.Error() == "invalid token" {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			msg := fmt.Sprintf("failed to verify new drive token: %v", err)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		newDrive, err := svc.UnmarshalDriveString(drvInfo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
			if err.Error() == "invalid token" {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
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
		ctx := context.WithValue(r.Context(), File, fileID)
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
		ctx := context.WithValue(r.Context(), Directory, dirID)
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
		ctx := context.WithValue(r.Context(), Drive, driveID)
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
		ctx := context.WithValue(r.Context(), User, userID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ErrorCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errMsg := chi.URLParam(r, "errMsg")
		if errMsg == "" {
			log.Print("errMsg is not set. will default to standard messaging")
			errMsg = "something's borked"
		}
		ctx := context.WithValue(r.Context(), Error, errMsg)
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
