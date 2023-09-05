package server

import (
	"net/http"

	"github.com/sfs/pkg/auth"

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

/*
AuthenticatedHanlder wrapper is to ensure a generic syntax
for ensuring all handlers that need user validation get it.

It basically extends the standard http.Handler function signature by
one argument (*auth.User).

Handlers must also have the signature:

	func(http.ResponseWriter, *http.Request, *auth.User)

so as to be wrappable by NewEnsureAuth() middleware

Examples:

	mux.Handle("/v1/users/", NewEnsureAuth(UsersHandler))
	mux.Handle("/v1/users/me", NewEnsureAuth(UsersMeHandler))
*/
type AuthenticatedHandler func(http.ResponseWriter, *http.Request, *auth.User)

type EnsureAuth struct {
	handler AuthenticatedHandler
}

// TODO:
func GetAuthenticatedUser(r *http.Request) (*auth.User, error) {
	// validate the session token in the request,
	// fetch the session state from the session store,
	// and return the authenticated user
	// or an error if the user is not authenticated

	return nil, nil
}

// calls GetAuthenticatedUser() prior to executing direct handler call
func (ea *EnsureAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		http.Error(w, "please sign-in", http.StatusUnauthorized)
		return
	}

	ea.handler(w, r, user)
}

func NewEnsureAuth(handlerToWrap AuthenticatedHandler) *EnsureAuth {
	return &EnsureAuth{handlerToWrap}
}

// ------- admin router --------------------------------

// A completely separate router for administrator routes
func adminRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(AdminOnly)
	// TODO: admin handlers
	// r.Get("/", adminIndex)
	// r.Get("/accounts", adminListAccounts)
	return r
}

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
