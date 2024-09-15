package client

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sfs/pkg/logger"
	"github.com/sfs/pkg/server"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

/*
This file defines a simple web client to server to facilitate a web UI for interacting with the main SFS server
in either a client or server management capacity
*/

// The web client is a local server that serves a web UI for
// interacting witht he local SFS client service.
type WebClient struct {
	log *logger.Logger
	svr *http.Server
}

func newWebClient(client *Client) *WebClient {
	return &WebClient{
		log: logger.NewLogger("Web Client", "None"),
		svr: &http.Server{
			Addr:         cfgs.Addr,
			Handler:      newWcRouter(client),
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
	}
}

func newWcRouter(client *Client) *chi.Mux {
	r := chi.NewRouter()

	// middlewares
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(server.EnableCORS)

	// serve js, css, and asset files
	r.Route("/static", func(r chi.Router) {
		fs := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
		r.Handle("/*", fs)
	})
	r.Route("/assets", func(r chi.Router) {
		fs := http.StripPrefix("/assets/", http.FileServer(http.Dir("assets")))
		r.Handle("/*", fs)
	})

	// home page
	r.Route("/", func(r chi.Router) {
		r.Get("/", client.HomePage)
	})

	// drive page
	r.Route("/drive", func(r chi.Router) {
		r.Get("/", client.DrivePage)
	})

	// error page
	r.Route("/error/{errMsg}", func(r chi.Router) {
		r.Use(server.ErrorCtx)
		r.Get("/", client.ErrorPage)
	})

	// search page
	r.Route("/search", func(r chi.Router) {
		r.Get("/", client.SearchPage)
		r.Post("/", client.SearchPage)
	})

	// user page
	r.Route("/user", func(r chi.Router) {
		r.Get("/", client.UserPage)
		r.Route("/edit", func(r chi.Router) {
			r.Get("/", client.EditInfo)
			r.Post("/", client.HandleNewUserInfo)
		})
		r.Route("/upload-pfp", func(r chi.Router) {
			r.Post("/", client.UpdatePfpHandler)
		})
		r.Route("/clear-pfp", func(r chi.Router) {
			r.Post("/", client.ClearPfpHandler)
		})
	})

	// add items to the service.
	r.Route("/add", func(r chi.Router) {
		r.Get("/", client.AddPage)
		r.Route("/new", func(r chi.Router) { // add in bulk using discover
			r.Post("/", client.AddItems)
		})
	})

	// emptying the recyle bin
	r.Route("/empty", func(r chi.Router) {
		r.Delete("/", client.EmptyRecycleBinHandler)
	})

	// upload files to server page
	r.Route("/upload", func(r chi.Router) {
		r.Get("/", client.UploadPage)
		r.Post("/", client.DropZoneHandler)
	})

	// settings page and handler
	r.Route("/settings", func(r chi.Router) {
		r.Get("/", client.SettingsPage)
		r.Post("/", client.SettingsHandler)
	})

	// file pages
	r.Route("/files", func(r chi.Router) {
		r.Route("/d/{fileID}", func(r chi.Router) {
			r.Use(server.FileCtx)
			r.Get("/", client.ServeFile) // get a copy of the file from the local client

		})
		r.Route("/i/{fileID}", func(r chi.Router) {
			r.Use(server.FileCtx)
			r.Get("/", client.FilePage) // get info about a file
			r.Route("/open-loc", func(r chi.Router) {
				r.Get("/", client.OpenFileLocHandler)
			})
		})
		r.Route("/delete", func(r chi.Router) {
			r.Delete("/", client.RemoveFileHandler)
		})
	})

	// dirs
	r.Route("/dirs", func(r chi.Router) {
		r.Route("/i/{dirID}", func(r chi.Router) {
			r.Use(server.DirCtx)
			r.Get("/", client.DirPage)
		})
		// r.Route("/d/{dirID}", func(r chi.Router) {
		// 	r.Use(server.DirCtx)
		// 	r.Get("/", client.DirPage)
		// })
	})

	// recycle bin page
	r.Route("/recycled", func(r chi.Router) {
		r.Get("/", client.RecycleBinPage)
	})

	return r
}

// run the web client and be able to shut it down with ctrl+c
func (wc *WebClient) start() {
	wc.log.Info("Starting web client...")
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		wc.log.Info("Shutting down web client...")
		shutdownCtx, _ := context.WithTimeout(serverCtx, 10*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				wc.log.Warn("Shutdown context deadline exceeded. Forcing exit...")
				if err := wc.svr.Close(); err != nil {
					log.Fatal(err)
				}
			}
		}()

		if err := wc.svr.Shutdown(shutdownCtx); err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	if err := wc.svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	<-serverCtx.Done()
}

// called by the CLI to start the web UI
func (c *Client) StartWebClient() {
	wc := newWebClient(c)
	// go func() {
	// 	time.Sleep(time.Millisecond * 500)
	// 	Openbrowser(homePage)
	// }()
	wc.start()
}
