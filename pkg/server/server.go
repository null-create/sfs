package server

import (
	"database/sql"
	"log"
	"net/http"
	"time"
)

type Server struct {
	StartTime time.Time

	Db  *sql.DB      // database connection
	Srv *http.Server // server
}

// NOTE: no handler(s) attached!
// Must be built and attached manually
func NewServer() *Server {

	// get http server configs
	c := SrvConfig()

	// TODO:
	// instantiate db

	// instantiate router
	rtr := NewRouter()

	return &Server{
		Srv: &http.Server{
			Handler:      rtr,
			Addr:         c.Server.Addr,
			ReadTimeout:  c.Server.TimeoutRead,
			WriteTimeout: c.Server.TimeoutWrite,
			IdleTimeout:  c.Server.TimeoutIdle,
		},
	}
}

// start the server
func (s *Server) Start() {
	s.StartTime = time.Now()
	if err := s.Srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[ERROR] server startup failed: %v", err)
	}
}

// shuts down server and returns the total run time
func (s *Server) Shutdown() float64 {
	if err := s.Srv.Close(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[ERROR] server shutdown failed: %v", err)
	}
	return s.RunTime()
}

// returns the current run time of the server in seconds
func (s *Server) RunTime() float64 {
	return time.Since(s.StartTime).Seconds()
}
