package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

// add json header to requests. added to middleware stack
// during router instantiation.
func ContentTypeJson(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json;charset=utf8")
		h.ServeHTTP(w, r)
	})
}

// func GetAuthenticatedUser(userID string, dbDir string, w http.ResponseWriter, r *http.Request) (*auth.User, error) {
// 	// TODO: validate the session token in the request

// 	// attempt to find data about the user from the the user db
// 	u, err := findUser(userID, dbDir)
// 	if err != nil {
// 		ServerErr(w, err.Error())
// 		return nil, err
// 	}
// 	return u, nil
// }

// get user info
func AuthUserHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// userID := chi.URLParam(r, "userID")
		// _, err := GetAuthenticatedUser(userID, "", w, r) // TODO get user db path
		// if err != nil {
		// 	ServerErr(w, "failed to get authenticated user")
		// 	return
		// }
		h.ServeHTTP(w, r)
	})
}

// check if this file is in the DB
func validateFile(w http.ResponseWriter, r *http.Request, fileID string) bool {
	f, err := findFile(fileID, getDBConn("Files"))
	if err != nil {
		msg := fmt.Sprintf("failed to find file (%s): ", fileID)
		w.Write([]byte(msg))
	}
	if f == nil { // not found
		return false
	}
	return true
}

func ValidateFile(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fileID := chi.URLParam(r, "fileID")
		if ok := validateFile(w, r, fileID); ok {
			h.ServeHTTP(w, r)
		} else {
			w.Write([]byte(fmt.Sprintf("file (%s) not found", fileID)))
		}
	})
}

// ------- admin router --------------------------------

// // A completely separate router for administrator routes
// func adminRouter() http.Handler {
// 	r := chi.NewRouter()
// 	r.Use(AdminOnly)
// 	// TODO: admin handlers
// 	// r.Get("/", adminIndex)
// 	// r.Get("/accounts", adminListAccounts)
// 	return r
// }

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
