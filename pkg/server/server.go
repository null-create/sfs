package server

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Server struct {
	StartTime time.Time
	Svr       *http.Server // http server
}

// instantiate a new HTTP server with an sfs service instance
func InitServer(newService bool, isAdmin bool) *Server {
	// load .env file
	BuildEnv()

	svr := ServerConfig() // get http server and service configs
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
	s.StartTime = time.Now()
	if err := s.Svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[ERROR] server startup failed: %v", err)
	}
}

// shuts down server and returns the total run time
func (s *Server) Shutdown() error {
	if err := s.Svr.Close(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server shutdown failed: %v", err)
	}
	return nil
}
