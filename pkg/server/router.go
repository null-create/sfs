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

GET    /v1/drive/{userID}             // "home". return a root directory listing

// ----- users (admin only)

GET     /v1/u/{userID}                // get info about a user
POST    /v1/u/{userID}                // create a new user
PUT     /v1/u/{userID}                // update a user
DELETE  /v1/u/{userID}                // delete a user

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
GET    /v1/u/{userID}/d/{dirID}      // download a .zip (or other compressed format) file of this directory and its contents
POST   /v1/u/{userID}/d/{dirID}      // create a directory on the server
UPDATE /v1/u/{userID}/d/{dirID}      // update a directory on the server
DELETE /v1/u/{userID}/d/{dirID}      // delete a directory on the server

// ----- sync operations

GET    /v1/u/{userID}/sync      // fetch file last sync times from server

POST   /v1/u/{userID}/sync      // send a last sync index object to the server
                                // generated from the local client directories to
								                // initiate a client/server file sync.
*/

// instantiate a new chi router
func NewRouter() *chi.Mux {
	// initialize API handlers
	api := NewAPI(isMode("NEW_SERVICE"), isMode("ADMIN_MODE"))

	// instantiate router
	r := chi.NewRouter()

	// chi's default middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// custom middleware
	r.Use(AuthUserHandler)
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
		// ----- files

		r.Get("/u/{userID}/f/files", api.GetFileInfo)      // get info about a file
		r.Get("/u/{userID}/f/{fileID}", api.GetFile)       // download a file from the server
		r.Post("/u/{userID}/f/{fileID}", api.PutFile)      // add a new file to the server
		r.Put("/u/{userID}/f/{fileID}", api.PutFile)       // update a file on the server
		r.Delete("/u/{userID}/f/{fileID}", api.DeleteFile) // delete a file on the server

		// ----- directories

		r.Get("/u/{userID}/d/dirs/", api.Placeholder)      // get list of user directories
		r.Delete("/u/{userID}/d/dirs/", api.Placeholder)   // delete all user directories
		r.Get("/u/{userID}/d/{dirID}/i", api.Placeholder)  // get info (file list) about a directory
		r.Get("/u/{userID}/d/{dirID}", api.Placeholder)    // download a .zip file of the directory from the server
		r.Post("/u/{userID}/d/{dirID}", api.Placeholder)   // create a (empty) directory to the server
		r.Put("/u/{userID}/d/{dirID}", api.Placeholder)    // update a directory on the server
		r.Delete("/u/{userID}/d/{dirID}", api.Placeholder) // delete a directory

		// ------- users

		// TODO: add some user API's (add/remove/search users)

		// ----- sync operations

		// fetch file last sync times for all
		// user files (in all directories) from server
		// to start a client side sync operation
		r.Get("/u/{userID}/sync", api.Placeholder)

		// send a newly generated last sync index to the
		// server to initiate a client/server file sync.
		r.Post("/u/{userID}/sync", api.Placeholder)
	})

	// Mount the admin sub-router
	// r.Mount("/admin", adminRouter)

	return r
}
