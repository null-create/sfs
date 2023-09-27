package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/sfs/pkg/auth"
)

// add json header to requests. added to middleware stack
// during router instantiation.
func ContentTypeJson(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json;charset=utf8")
		h.ServeHTTP(w, r)
	})
}

func AuthenticateUser(w http.ResponseWriter, r *http.Request) (*auth.User, error) {
	tok := auth.NewT()

	// retrieve jwt token from request & verify
	reqToken, err := tok.Extract(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}
	userID, err := tok.Verify(reqToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, err
	}

	// attempt to find data about the user from the the user db
	u, err := findUser(userID, getDBConn("Users"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	} else if u == nil {
		http.Error(w, fmt.Sprintf("user (id=%s) not found", userID), http.StatusNotFound)
		return nil, err
	}
	return u, nil
}

// get user info
func AuthUserHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := AuthenticateUser(w, r)
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
		file, err := findFile(chi.URLParam(r, "fileID"), getDBConn("Users"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if file == nil {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), "file", file)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func DriveCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		drive, err := findDrive(chi.URLParam(r, "driveID"), getDBConn("Drives"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if drive == nil {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), "drive", drive)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func DirCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dir, err := findDir(chi.URLParam(r, "driveID"), getDBConn("Directories"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if dir == nil {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), "directory", dir)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserCtx(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := findUser(chi.URLParam(r, "driveID"), getDBConn("Users"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if user == nil {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), "user", user)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// func AdminOnly(h http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		ctx := r.Context()
// 		_, ok := ctx.Value("acl.permission").(float64)
// 		if !ok {
// 			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
// 			return
// 		}
// 		h.ServeHTTP(w, r)
// 	})
// }
