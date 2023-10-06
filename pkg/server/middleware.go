package server

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/sfs/pkg/auth"

	"github.com/go-chi/chi"
)

type Context string

// context enums
const (
	File      Context = "file"
	Directory Context = "directory"
	Drive     Context = "drive"
	User      Context = "user"
)

// add json header to requests. added to middleware stack
// during router instantiation.
func ContentTypeJson(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json;charset=utf8")
		h.ServeHTTP(w, r)
	})
}

func NewUser(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := ServiceConfig()
		newUserID := auth.NewUUID()
		newDriveRoot := filepath.Join(c.S.SvcRoot, "users", newUserID)

		ctx := r.Context()
		newUserName := ctx.Value("userName").(string)
		newUserAlias := ctx.Value("userAlias").(string)
		newUserEmail := ctx.Value("userEmail").(string)

		newUser := auth.NewUser(
			newUserName, newUserAlias, newUserEmail,
			newUserID, newDriveRoot, false,
		)
		// this basically just repackages the previous
		// context with a new user object
		newCtx := context.WithValue(ctx, User, newUser)

		h.ServeHTTP(w, r.WithContext(newCtx))
	})
}

// retrieve jwt token from request & verify
func AuthenticateUser(authReq string) (*auth.User, error) {
	// verify request token
	tok := auth.NewT()
	reqToken, err := tok.Extract(authReq)
	if err != nil {
		return nil, err
	}
	userID, err := tok.Verify(reqToken)
	if err != nil {
		return nil, err
	}

	// attempt to find data about the user from the the user db
	u, err := findUser(userID, getDBConn("Users"))
	if err != nil {
		return nil, err
	} else if u == nil {
		return nil, fmt.Errorf("user (id=%s) not found", userID)
	}
	return u, nil
}

// get user info
func AuthUserHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authReq := r.Header.Get("Authorization")
		if authReq == "" {
			http.Error(w, "header had no request token", http.StatusBadRequest)
			return
		}
		_, err := AuthenticateUser(authReq)
		if err != nil {
			http.Error(w, "failed to get authenticated user", http.StatusInternalServerError)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// ------ context --------------------------------

func FileCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fileID := chi.URLParam(r, "fileID")
		file, err := findFile(fileID, getDBConn("Files"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if file == nil {
			http.Error(w, fmt.Sprintf("file (id=%s) not found", fileID), http.StatusNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), File, file)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func DriveCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		driveID := chi.URLParam(r, "driveID")
		drive, err := findDrive(driveID, getDBConn("Drives"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if drive == nil {
			http.Error(w, fmt.Sprintf("drive (id=%s) not found", driveID), http.StatusNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), Drive, drive)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func DirCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dirID := chi.URLParam(r, "dirID")
		dir, err := findDir(dirID, getDBConn("Directories"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if dir == nil {
			http.Error(w, fmt.Sprintf("directory (id=%s) not found", dirID), http.StatusNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), Directory, dir)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")
		user, err := findUser(userID, getDBConn("Users"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if user == nil {
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
