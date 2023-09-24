package server

import (
	"fmt"
	"log"
	"net/http"
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
	BuildEnv()            // initialize environment variables
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

// start the server
func (s *Server) Start() {
	s.StartTime = time.Now().UTC()
	log.Printf("starting server...")
	if err := s.Svr.ListenAndServe(); err != nil {
		log.Fatalf("[ERROR] server startup failed: %v", err)
	}
}

// shuts down server and returns the total run time
func (s *Server) Shutdown() error {
	log.Printf("shutting down server...")
	if err := s.Svr.Close(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server shutdown failed: %v", err)
	}
	return nil
}
