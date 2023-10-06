package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

/*
ROUTES:

// ----- meta

GET     /v1/drive/{userID}        // "home". return a root directory listing

// ----- users (admin only)

POST    /v1/users/{userID}/new   // create a new user
GET     /v1/users/{userID}       // get info about a user
PUT     /v1/users/{userID}       // update a user
DELETE  /v1/users/{userID}       // delete a user

// ----- files

GET    /v1/i/files/{fileID}    // get info about a file
GET    /v1/files/{fileID}      // download a file from the server
POST   /v1/files/{fileID}      // send a new file to the server
PUT    /v1/files/{fileID}      // update a file on the server
DELETE /v1/files/{fileID}      // delete a file on the server

// ---- directories

GET    /v1/i/dirs/{dirID}    // get list of files and subdirectories for this directory
GET    /v1/dirs/{dirID}      // download a .zip (or other compressed format) file of this directory and its contents
POST   /v1/dirs/{dirID}      // create a directory on the server
PUT    /v1/dirs/{dirID}      // update a directory on the server
DELETE /v1/dirs/{dirID}      // delete a directory on the server

// ----- sync operations

GET    /v1/users/{userID}/sync  // fetch file last sync times from server

POST   /v1/users/{userID}/sync  // send a last sync index object to the server
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
	// r.Use(AuthUserHandler)

	r.Use(render.SetContentType(render.ContentTypeJSON))

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(time.Minute))

	// placeholder for sfs "homepage"
	// this will eventually display a simple service index page
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})

	// TODO: rework so ctx middleware isn't used when creating
	// files/users/directories. only need ctx for existing items.

	//v1 routing
	r.Route("/v1", func(r chi.Router) {
		r.Route("/files", func(r chi.Router) {
			r.Route("/{fileID}", func(r chi.Router) {
				r.Use(FileCtx)
				r.Get("/", api.Placeholder)    // get a file from the server
				r.Post("/", api.Placeholder)   // add a new file to the server
				r.Put("/", api.Placeholder)    // update a file on the server
				r.Delete("/", api.Placeholder) // delete a file on the server
			})
			r.Route("/i/{fileID}", func(r chi.Router) {
				r.Use(FileCtx)
				r.Get("/", api.Placeholder) // get info about a file
			})
		})

		r.Route("/dirs", func(r chi.Router) {
			r.Route("/{dirID}", func(r chi.Router) {
				r.Use(DirCtx)
				r.Get("/", api.Placeholder)    // get a directory as a zip file
				r.Post("/", api.Placeholder)   // create a (empty) directory to the server
				r.Put("/", api.Placeholder)    // update a directory on the server
				r.Delete("/", api.Placeholder) // delete a directory
			})
			r.Route("/i/{dirID}", func(r chi.Router) {
				r.Use(DirCtx)
				r.Get("/", api.Placeholder) // get info about a directory
			})
		})

		r.Route("/drive", func(r chi.Router) {
			r.Route("/{driveID}", func(r chi.Router) {
				r.Use(DriveCtx)
				r.Get("/", api.Placeholder) // "home" page for files.
			})
		})

		// sync operations
		r.Route("/users", func(r chi.Router) {
			r.Route("/{userID}", func(r chi.Router) {
				r.Use(UserCtx)
				r.Route("/sync", func(r chi.Router) {
					// fetch file last sync times for all
					// user files (in all directories) from server
					// to start a client side sync operation
					r.Get("/", api.Placeholder)
					// send a newly generated last sync index to the
					// server to initiate a client/server file sync.
					r.Post("/", api.Placeholder)
				})
			})
		})
	})

	// :)
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	// mount the admin sub-router
	r.Mount("/admin", adminRouter())

	// generates a json document of our routing
	// fmt.Println(docgen.MarkdownRoutesDoc(r, docgen.MarkdownOpts{
	// 	ProjectPath: "github.com/go-chi/chi/v5",
	// 	Intro:       "Welcome to the chi/_examples/rest generated docs.",
	// }))

	return r
}

// ------- admin router --------------------------------

// A completely separate router for administrator routes
func adminRouter() http.Handler {
	r := chi.NewRouter()

	// r.Use(AdminOnly)

	// initialize API handlers
	api := NewAPI(isMode("NEW_SERVICE"), true)

	r.Route("/users", func(r chi.Router) {
		r.Route("/{userID}", func(r chi.Router) {
			r.Use(UserCtx)
			r.Get("/", api.Placeholder)    // get info about a user
			r.Put("/", api.Placeholder)    // update a user
			r.Delete("/", api.Placeholder) // delete a user)
		})
		r.Route("/new", func(r chi.Router) {
			r.Use(NewUser)
			r.Post("/", api.Placeholder) // add a new user
		})
	})
	return r
}
