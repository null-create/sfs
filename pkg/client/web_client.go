package client

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
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
	log *logger.Logger
	svr *http.Server
}

func newWebClient(client *Client) *WebClient {
	return &WebClient{
		log: logger.NewLogger("Web Client", "None"),
		svr: &http.Server{
			Addr:         cfgs.Addr,
			Handler:      handlers(client),
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		}}
}

// get the absolute path to each template we'll need to render.
// go's template module seems to only work with absolute paths.
func setupTemplatePaths() (map[string]string, error) {
	var templates = []string{
		"./pkg/client/views/templates/index.html",
		"./pkg/client/views/templates/file.html",
		"./pkg/client/views/templates/folder.html",
		"./pkg/client/views/templates/frame.html",
		"./pkg/client/views/templates/toolbar.html",
		"./pkg/client/views/templates/header.html",
	}
	var paths = make(map[string]string)
	for _, tpath := range templates {
		var tp, err = filepath.Abs(tpath)
		if err != nil {
			return nil, err
		}
		paths[filepath.Base(tpath)] = tp
	}
	return paths, nil
}

func handlers(client *Client) *chi.Mux {
	rtr := chi.NewRouter()

	// middlewares
	rtr.Use(middleware.Logger)
	rtr.Use(middleware.Recoverer)
	rtr.Use(server.EnableCORS)

	// get paths to HTML tempaltes and parse each one
	templatePaths, err := setupTemplatePaths()
	if err != nil {
		client.log.Error(fmt.Sprintf("failed to build html template paths: %v", err))
		os.Exit(1)
	}
	tmpl, err := template.ParseFiles(
		templatePaths["index.html"],
		templatePaths["frame.html"],
		templatePaths["toolbar.html"],
		templatePaths["header.html"],
	)

	// serve static css and asset files
	rtr.Get("/static", func(w http.ResponseWriter, r *http.Request) {
		fs := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
		rtr.Handle("/*", fs)
	})
	rtr.Get("/assets", func(w http.ResponseWriter, r *http.Request) {
		fs := http.StripPrefix("/assets/", http.FileServer(http.Dir("assets")))
		rtr.Handle("/*", fs)
	})

	// home page w/all items
	rtr.Get("/", func(w http.ResponseWriter, r *http.Request) {
		indexData := Index{
			Files: client.Drive.GetFiles(),
			Dirs:  client.Drive.GetDirs(),
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = tmpl.Execute(w, indexData)
		if err != nil {
			client.log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// // individual file page
	// rtr.Get("/{fileID}", func(w http.ResponseWriter, r *http.Request) {

	// })

	// // all folders page
	// rtr.Get("/directories", func(w http.ResponseWriter, r *http.Request) {

	// })

	// // individual folder page
	// rtr.Get("/{dirID}", func(w http.ResponseWriter, r *http.Request) {

	// })

	// // upload page
	// rtr.Get("/upload", func(w http.ResponseWriter, r *http.Request) {

	// })

	// download pages

	return rtr
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		log.Fatal("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
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
	// 	openbrowser(homePage)
	// }()
	wc.start()
}

// open the web client home page in a new browser window
func openbrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	exec.Command(cmd, args...).Start()
}
