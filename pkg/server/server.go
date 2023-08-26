package server

import (
	"log"
	"net/http"
	"time"

	"github.com/sfs/pkg/db"
)

type Server struct {
	StartTime time.Time

	Svc *Service     // SFS service instance
	Db  *db.Query    // database connection
	Srv *http.Server // http server
}

// Instantiate a new HTTP server with an sfs service instance
func NewServer(newService bool, isAdmin bool) *Server {

	// get http server and service configs
	c := SrvConfig()
	srvConf := GetServiceConfig()

	// instantiate service
	svc, err := Init(newService, isAdmin)
	if err != nil {
		log.Fatalf("[ERROR] failed to initialize new service instance: %v", err)
	}

	// instantiate router
	rtr := NewRouter()

	return &Server{
		Svc: svc,
		Db:  db.NewQuery(srvConf.ServiceRoot),
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
