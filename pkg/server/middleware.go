package server

import (
	"fmt"
	"net/http"
	"strings"

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
	// retrieve jwt token from request
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer")
	if len(splitToken) != 2 { // bearer token not in proper format
		http.Error(w, "invalid token format", http.StatusBadRequest)
	}
	reqToken = strings.TrimSpace(splitToken[1])

	// verify token
	tok := auth.NewT()
	userID, err := tok.Verify(reqToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	// attempt to find data about the user from the the user db
	u, err := findUser(userID, getDBConn("Users"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if u == nil {
		http.Error(w, fmt.Sprintf("user (id=%s) not found", userID), http.StatusNotFound)
	}
	return u, nil
}

// get user info
func AuthUserHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := AuthenticateUser(w, r)
		if err != nil {
			ServerErr(w, "failed to get authenticated user")
			return
		}
		h.ServeHTTP(w, r)
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
