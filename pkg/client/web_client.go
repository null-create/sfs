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

var homePage = "http://" + cfgs.Addr

// The web client is a local server that serves a web UI for
// interacting witht he local SFS client service.
type WebClient struct {
	client *Client
	log    *logger.Logger
	svr    *http.Server
}

func newWebClient(client *Client) *WebClient {
	return &WebClient{
		log:    logger.NewLogger("Web Client", "None"),
		client: client,
		svr: &http.Server{
			Addr:         cfgs.Addr,
			Handler:      newWcRouter(client),
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		}}
}

func newWcRouter(client *Client) *chi.Mux {
	rtr := chi.NewRouter()

	// middlewares
	rtr.Use(middleware.Logger)
	rtr.Use(middleware.Recoverer)
	rtr.Use(server.EnableCORS)

	// serve css and asset files
	rtr.Route("/static", func(rtr chi.Router) {
		fs := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
		rtr.Handle("/*", fs)
	})
	rtr.Route("/assets", func(rtr chi.Router) {
		fs := http.StripPrefix("/assets/", http.FileServer(http.Dir("assets")))
		rtr.Handle("/*", fs)
	})

	// home page w/all items
	rtr.Route("/", func(rtr chi.Router) {
		rtr.Get("/", client.HomePage)
	})

	// error page
	rtr.Route("/error", func(rtr chi.Router) {
		rtr.Get("/", client.ErrorPage)
	})

	// files
	rtr.Route("/files", func(rtr chi.Router) {
		rtr.Route("/d/{fileID}", func(rtr chi.Router) {
			rtr.Use(server.FileCtx)
			rtr.Get("/", client.ServeFile) // get a copy of the file from the local client

		})
		rtr.Route("/i/{fileID}", func(rtr chi.Router) {
			rtr.Use(server.FileCtx)
			rtr.Get("/", client.FilePage) // get info about a file
		})
		rtr.Post("/new", client.NewFile) // add a new file to the service
	})

	// dirs
	rtr.Route("/dirs", func(rtr chi.Router) {
		rtr.Route("/i/{dirID}", func(rtr chi.Router) {
			rtr.Use(server.DirCtx)
			rtr.Get("/", client.DirPage)
		})
		// rtr.Route("/d/{dirID}", func(rtr chi.Router) {
		// 	rtr.Use(server.DirCtx)
		// 	rtr.Get("/", client.DirPage)
		// })
	})

	// // upload page
	// rtr.Get("/upload", func(w http.ResponseWriter, r *http.Request) {

	// })

	// download pages

	return rtr
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
