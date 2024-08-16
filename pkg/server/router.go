package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

/*
ROUTES:

// ----- meta

GET     /v1/drive/{userID}        // "home". return a root directory listing

// ----- users (admin only)

GET     /v1/users/{userID}       // get info about a user
POST    /v1/users/new            // create a new user
PUT     /v1/users/{userID}       // update a user
DELETE  /v1/users/{userID}       // delete a user

// ----- files

GET    /v1/files/{fileID}/i    // get info about a file
POST   /v1/files/new           // send a new file to the server
GET    /v1/files/{fileID}      // download a file from the server
PUT    /v1/files/{fileID}      // update a file on the server
DELETE /v1/files/{fileID}      // delete a file on the server

// ---- directories

GET    /v1/i/dirs/{dirID}    // get list of files and subdirectories for this directory
POST   /v1/dirs/new          // create a directory on the server
GET    /v1/dirs/{dirID}      // download a .zip (or other compressed format) file of this directory and its contents
PUT    /v1/dirs/{dirID}      // update a directory on the server
DELETE /v1/dirs/{dirID}      // delete a directory on the server

// ----- sync operations

GET    /v1/sync/{driveID}    // fetch file last sync times from server
POST   /v1/sync/{driveID}    // send a last sync index object to the server
                             // generated from the local client directories to
								             // initiate a client/server file sync.
*/

// instantiate a new chi router
func NewRouter() *chi.Mux {
	// initialize API handlers and SFS service instance
	api := NewAPI(svcCfg.NewService, svcCfg.IsAdmin)

	// instantiate router
	r := chi.NewRouter()

	// standard middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(time.Minute))

	// custom middleware
	// r.Use(AuthUserHandler)
	r.Use(ContentTypeJson) // will be overridden by streaming API endpoints
	r.Use(EnableCORS)      // used for working with the client web interface

	// placeholder for sfs "homepage"
	// this will eventually display a simple service index page
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})

	//v1 routing
	r.Route("/v1", func(r chi.Router) {
		// Get the total runtime of the server and SFS service
		r.Route("/runtime", func(r chi.Router) {
			r.Get("/", api.GetRunTime)
		})
		r.Route("/users", func(r chi.Router) {
			r.Route("/{userID}", func(r chi.Router) {
				r.Use(UserCtx)
				r.Get("/", api.GetUser)       // get info about a user
				r.Put("/", api.UpdateUser)    // update a user
				r.Delete("/", api.DeleteUser) // delete a user
			})
			r.Route("/new", func(r chi.Router) {
				r.Use(NewUserCtx)
				r.Post("/", api.AddNewUser) // add a new user
			})
			r.Route("/all", func(r chi.Router) {
				// TODO: add admin-only context here
				// get a list of all active users
				r.Get("/", api.GetAllUsers)
			})
		})

		// files
		r.Route("/files", func(r chi.Router) {
			r.Route("/{fileID}", func(r chi.Router) {
				r.Use(FileCtx)
				r.Get("/", api.ServeFile)     // get a file from the server
				r.Put("/", api.PutFile)       // update a file on the server
				r.Delete("/", api.DeleteFile) // delete a file on the server
			})
			r.Route("/i/all/{userID}", func(r chi.Router) {
				r.Use(AllUsersFilesCtx)
				r.Get("/", api.GetAllFileInfo) // get info about all user-specific files
			})
			r.Route("/new", func(r chi.Router) { // add a new file on the server
				r.Use(NewFileCtx)
				r.Post("/", api.PutFile)
			})
			r.Route("/i/{fileID}", func(r chi.Router) {
				r.Use(FileCtx)
				r.Get("/", api.GetFileInfo) // get info about a file
			})
			// temp for testing
			r.Route("/all", func(r chi.Router) {
				r.Get("/", api.GetAllFileInfo) // get info
			})
		})

		// directories
		// NOTE: Directories are not supported at this time, but we'll keep these
		// endpoints in place for future iterations.
		r.Route("/dirs", func(r chi.Router) {
			// specific directories
			r.Route("/{dirID}", func(r chi.Router) {
				r.Use(DirCtx)
				r.Get("/", api.GetDir)       // get a directory as a zip file
				r.Put("/", api.PutDir)       // update a directory on the server by sending a zip file and unpacking
				r.Delete("/", api.DeleteDir) // delete a directory
			})
			// create a new directory
			r.Route("/new", func(r chi.Router) {
				r.Use(NewDirectoryCtx)
				r.Post("/", api.NewDir)
			})
			// get info about a directory
			r.Route("/i/{dirID}", func(r chi.Router) {
				r.Use(DirCtx)
				r.Get("/", api.GetDirInfo)
			})
			// get info about *all* directories
			r.Route("/i/all/{userID}", func(r chi.Router) {
				r.Use(UserCtx)
				r.Get("/", api.GetUsersDirs)
			})
			// temp for testing
			r.Route("/all", func(r chi.Router) {
				r.Get("/", api.GetAllDirsInfo)
			})
		})

		// drives
		r.Route("/drive/{driveID}", func(r chi.Router) {
			r.Use(DriveCtx)
			r.Get("/", api.GetDrive) // "home" page data for all user's files, directories, etc.
			// NOTE: new drives are created when a new user is added.
		})
		// add a new drive
		r.Route("/drive/new", func(r chi.Router) {
			r.Use(NewDriveCtx)
			r.Post("/", api.NewDrive)
		})

		// sync operations
		r.Route("/sync/{driveID}", func(r chi.Router) {
			r.Use(DriveCtx)
			// fetch file last sync times for all
			// user files (in all directories) from server
			r.Get("/", api.GetIdx)
			// generate a new sync index for all files on the server
			// for this user. Populates LastSync map in index.
			r.Get("/index", api.GenIndex)
			// refreshes a drives ToUpdate map (assumes LastSync is current),
			// and returns the servers sync index for this drive/user
			r.Get("/update", api.GetUpdates)
		})
	})

	// :)
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	// mount the admin sub-router
	// r.Mount("/admin", adminRouter())

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
	api := NewAPI(svcCfg.NewService, true)

	r.Route("/users", func(r chi.Router) {
		r.Route("/{userID}", func(r chi.Router) {
			r.Use(UserCtx)
			r.Get("/", api.GetUser)        // get info about a user
			r.Put("/", api.Placeholder)    // update a user
			r.Delete("/", api.Placeholder) // delete a user)

		})
		r.Route("/new", func(r chi.Router) {
			r.Use(NewUserCtx)
			r.Post("/", api.Placeholder) // add a new user
		})
		r.Route("/all", func(r chi.Router) {
			// get a list of all active users
			r.Get("/", api.GetAllUsers)
		})
	})

	// TODO: other db query routes.

	return r
}
