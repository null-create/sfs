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

GET    /v1/u/{userID}                 // "home". return a root directory listing

// ----- files

GET    /v1/u/{userID}/f/files         // get list of user files and directories
POST   /v1/u/{userID}/f/files         // send a file to the server

GET    /v1/u/{userID}/f/{fileID}/i/   // get info about a file
GET    /v1/u/{userID}/f/{fileID}      // download a file from the server
POST   /v1/u/{userID}/f/{fileID}      // send a new file to the server
UPDATE /v1/u/{userID}/f/{fileID}      // update a file on the server
DELETE /v1/u/{userID}/f/{fileID}      // delete a file on the server

// ---- directories

GET    /v1/u/{userID}/d/dirs         // get list of user directories

GET    /v1/u/{userID}/d/{dirID}/i/   // get list of files and subdirectories for this directory
GET    /v1/u/{userID}/d/{dirID}      // download a .zip file from the server of this directory
POST   /v1/u/{userID}/d/{dirID}      // create a directory to the server
UPDATE /v1/u/{userID}/d/{dirID}      // update a directory on the server
DELETE /v1/u/{userID}/d/{dirID}      // delete a directory

// ----- sync operations

GET    /v1/u/{userID}/sync      // fetch file last sync times from server

POST   /v1/u/{userID}/sync      // send a last sync index object to the server
                                // generated from the local client directories to
																// initiate a client/server file sync.
*/

// instantiate a new chi router
func NewRouter() *chi.Mux {

	// instantiate router
	r := chi.NewRouter()

	// add default middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(ContentTypeJson)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(time.Minute))

	// placeholder for sfs "homepage"
	// this will eventually display a simple service index page
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})

	// TODO: add handlers

	//v1 routing
	r.Route("/v1", func(r chi.Router) {
		//  get user info
		r.Get("/u/{userID}", nil)

		// ----- files

		// get list of user files and directories
		r.Get("/u/{userID}/f/files", nil)
		// get info about a file
		r.Get("/u/{userID}/f/{fileID}/i", nil)
		// download a file from the server
		r.Get("/u/{userID}/f/{fileID}", nil)
		// add a new file to the server
		r.Post("/u/{userID}/f/{fileID}", nil)
		// update a file on the server
		r.Put("/u/{userID}/f/{fileID}", nil)
		// delete a file on the server
		r.Delete("/u/{userID}/f/{fileID}", nil)

		// ----- directories

		// get list of user directories
		r.Get("/u/{userID}/d/dirs/", nil)
		// delete all user directories
		r.Delete("/u/{userID}/d/dirs/", nil)

		// get info (file list) about a directory
		r.Get("/u/{userID}/d/{dirID}/i", nil)
		// download a .zip file of the directory from the server
		r.Get("/u/{userID}/d/{dirID}", nil)
		// create a (empty) directory to the server
		r.Post("/u/{userID}/d/{dirID}", nil)
		// update a directory on the server
		r.Put("/u/{userID}/d/{dirID}", nil)
		// delete a directory
		r.Delete("/u/{userID}/d/{dirID}", nil)

		// ----- sync operations

		// fetch file last sync times for all
		// user files (in all directories) from server
		// to start a client side sync operation
		r.Get("/u/{userID}/sync", nil)

		// send a newly generated last sync index to the
		// server to initiate a client/server file sync.
		r.Post("/u/{userID}/sync", nil)
	})

	// Mount the admin sub-router
	// r.Mount("/admin", adminRouter())

	return r
}
