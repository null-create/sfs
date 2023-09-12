package server

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	StartTime time.Time

	Svc *Service     // SFS service instance
	Svr *http.Server // http server
}

// instantiate a new HTTP server with an sfs service instance
func NewServer(newService bool, isAdmin bool) *Server {
	// load env variables from .env file
	BuildEnv()

	// initialize sfs service
	svc, err := Init(newService, isAdmin)
	if err != nil {
		log.Fatalf("[ERROR] failed to initialize new service instance: %v", err)
	}
	svr := ServerConfig() // get http server and service configs
	rtr := NewRouter()    // instantiate router

	return &Server{
		Svc: svc,
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
// func (s *Server) Start() {
// 	s.StartTime = time.Now()
// 	if err := s.Svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
// 		log.Fatalf("[ERROR] server startup failed: %v", err)
// 	}
// }

// start the server in a separate goroutine
func (s *Server) Start(wg *sync.WaitGroup) {
	s.StartTime = time.Now()
	go func() {
		defer wg.Done() // let main know we are done cleaning up
		if err := s.Svr.ListenAndServe(); err != nil {
			log.Fatalf("[ERROR] failed to launch server: %v", err)
		}
	}()
}

// shuts down server and returns the total run time
func (s *Server) Shutdown() error {
	if err := s.Svr.Close(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server shutdown failed: %v", err)
	}
	return nil
}
