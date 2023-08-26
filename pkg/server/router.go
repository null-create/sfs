package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

/*
ROUTES:

// ----- meta

GET    /v1/u/{userID}                // "home". return a root directory listing

// ----- files

GET    /v1/u/{userID}/f/files       // get list of user files and directories
POST   /v1/u/{userID}/f/files       // send a file to the server

GET    /v1/u/{userID}/f/i/{fileID}  // get info about a file
GET    /v1/u/{userID}/f/{fileID}    // download a file from the server
POST   /v1/u/{userID}/f/{fileID}    // send a new file to the server
UPDATE /v1/u/{userID}/f/{fileID}    // update a file on the server
DELETE /v1/u/{userID}/f/{fileID}    // delete a file on the server

// ---- directories

GET    /v1/u/{userID}/d/dirs        // get list of user directories

GET    /v1/u/{userID}/d/i/{dirID}   // get list of files and subdirectories for this directory
GET    /v1/u/{userID}/d/{dirID}     // download a .zip file from the server of this directory
POST   /v1/u/{userID}/d/{dirID}     // create a directory to the server
UPDATE /v1/u/{userID}/d/{dirID}     // update a directory on the server
DELETE /v1/u/{userID}/d/{dirID}     // delete a directory

// ----- sync operations

GET    /v1/u/{userID}/sync      // fetch file last sync times from server

POST   /v1/u/{userID}/sync      // send a last sync index object to the server
                                // generated from the local client directories to
																// initiate a client/server file sync.
*/

// add json header to requests
func ContentTypeJson(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json;charset=utf8")
		next.ServeHTTP(w, r)
	})
}

// get a URL parameter value from a request
func GetParam(req *http.Request, param string) string {
	return chi.URLParam(req, param)
}

// instantiate a new chi router
func NewRouter() *chi.Mux {

	// instantiate router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(ContentTypeJson)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	// placeholder for homepage
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})

	// TODO: add handlers

	//v1 routing
	r.Route("/v1", func(r chi.Router) {
		// "home". return a root directory listing
		r.Get("/u/{userID}", nil)

		// ----- files

		r.Get("/u/{userID}/f/files", nil) // get list of user files and directories

		r.Get("/u/{userID}/f/i/{fileID}", nil)  // get info about a file
		r.Get("/u/{userID}/f/{fileID}", nil)    // download a file from the server
		r.Post("/u/{userID}/f/{fileID}", nil)   // add a new file to the server
		r.Put("/u/{userID}/f/{fileID}", nil)    // update a file on the server
		r.Delete("/u/{userID}/f/{fileID}", nil) // delete a file on the server

		// ----- directories

		r.Get("/u/{userID}/d/dirs/", nil)    // get list of user directories
		r.Delete("/u/{userID}/d/dirs/", nil) // delete all user directories

		r.Get("/u/{userID}/d/i/{dirID}", nil)  // get info (file list) about a directory
		r.Get("/u/{userID}/d/{dirID}", nil)    // download a .zip file of the directory from the server
		r.Post("/u/{userID}/d/{dirID}", nil)   // create a (empty) directory to the server
		r.Put("/u/{userID}/d/{dirID}", nil)    // update a directory on the server
		r.Delete("/u/{userID}/d/{dirID}", nil) // delete a directory

		// ----- sync operations

		// fetch file last sync times for all
		// user files (in all directories) from server
		r.Get("/u/{userID}/sync", nil)

		// send a last sync index to the server to
		// initiate a client/server file sync.
		r.Post("/u/{userID}/sync", nil)
	})

	// Mount the admin sub-router
	r.Mount("/admin", adminRouter())

	return r
}

// A completely separate router for administrator routes
func adminRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(AdminOnly)
	// r.Get("/", adminIndex)
	// r.Get("/accounts", adminListAccounts)
	return r
}

func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, ok := ctx.Value("acl.permission").(float64)
		if !ok {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
