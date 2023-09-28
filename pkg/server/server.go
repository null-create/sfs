package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	StartTime time.Time
	Svr       *http.Server
}

// instantiate a new HTTP server with an sfs service instance
//
// establishes environment variables and intializes a new router
// with sfs handlers
func NewServer() *Server {
	svr := ServerConfig() // get server configs
	rtr := NewRouter()    // instantiate router

	return &Server{
		Svr: &http.Server{
			Handler:      rtr,
			Addr:         svr.Server.Addr,
			ReadTimeout:  svr.Server.TimeoutRead,
			WriteTimeout: svr.Server.TimeoutWrite,
			IdleTimeout:  svr.Server.TimeoutIdle,
		},
	}
}

// returns the current run time of the server in seconds
func (s *Server) RunTime() float64 {
	return time.Since(s.StartTime).Seconds()
}

// forcibly shuts down server and returns total run time
func (s *Server) Shutdown() (float64, error) {
	log.Printf("forcing server shut down...")
	if err := s.Svr.Close(); err != nil && err != http.ErrServerClosed {
		return 0, fmt.Errorf("server shutdown failed: %v", err)
	}
	return s.RunTime(), nil
}

// runs server with graceful shutdowns
//
// based off examples from chi
func (s *Server) Run() {
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// shutdown signal with grace period of 10 seconds
		shutdownCtx, _ := context.WithTimeout(serverCtx, 10*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Print("shutdown timed out. forcing exit.")
				if _, err := s.Shutdown(); err != nil {
					log.Fatal(err)
				}
			}
		}()

		log.Printf("shutting down server...")
		err := s.Svr.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	log.Printf("starting server...")
	err := s.Svr.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	<-serverCtx.Done()
}
