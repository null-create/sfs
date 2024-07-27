package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/logger"
)

type Server struct {
	StartTime time.Time
	Svr       *http.Server
	log       *logger.Logger
}

// instantiate a new HTTP server with an sfs service instance
// contained within the router
// TODO: more germaine configurations so the server can handle
// a large amount of active connections and requests.
func NewServer() *Server {
	return &Server{
		StartTime: time.Now().UTC(),
		log:       logger.NewLogger("Server", auth.NewUUID()),
		Svr: &http.Server{
			// NewRouter() instantiates the server-side SFS service instance
			// and handles client requests.
			Handler: NewRouter(),
			// server configs
			Addr:         svrCfg.Addr,
			ReadTimeout:  svrCfg.TimeoutRead,
			WriteTimeout: svrCfg.TimeoutWrite,
			IdleTimeout:  svrCfg.TimeoutIdle,
			ConnState: func(n net.Conn, h http.ConnState) {
				// TODO: handle when a client state is idle or hijacked.
			},
			// TLS configs?
		},
	}
}

// returns the current run time of the server
// as a HH:MM:SS formatted string.
func (s *Server) RunTime() string {
	return secondsToTimeStr(time.Since(s.StartTime).Seconds())
}

// forcibly shuts down server and returns total run time in seconds.
func (s *Server) Shutdown() (string, error) {
	if err := s.Svr.Close(); err != nil && err != http.ErrServerClosed {
		return "0", fmt.Errorf("server shutdown failed: %v", err)
	}
	return s.RunTime(), nil
}

// starts a server that can be shut down via ctrl-c
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
				s.log.Warn("shutdown timed out. forcing exit.")
				if _, err := s.Shutdown(); err != nil {
					log.Fatal(err)
				}
				s.log.Info(fmt.Sprintf("server run time: %s", s.RunTime()))
			}
		}()

		s.log.Info("shutting down server...")
		if err := s.Svr.Shutdown(shutdownCtx); err != nil {
			log.Fatal(err)
		}
		s.log.Info(fmt.Sprintf("server run time: %v", s.RunTime()))
		serverStopCtx()
	}()

	s.log.Info("starting server...")
	if err := s.Svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	<-serverCtx.Done()
}

// start a server that can be shut down using a shutDown bool channel.
func (s *Server) Start(shutDown chan bool) {
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	go func() {
		// blocks until shutDown = true
		// (set by outer test and passed after checks are completed (or failed))
		<-shutDown

		// shutdown signal with grace period of 10 seconds
		shutdownCtx, _ := context.WithTimeout(serverCtx, 10*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				s.log.Error("shutdown timed out. forcing exit.")
				if _, err := s.Shutdown(); err != nil {
					log.Fatal(err)
				}
				s.log.Info(fmt.Sprintf("server run time: %s", s.RunTime()))
			}
		}()

		s.log.Info("shutting down server...")
		err := s.Svr.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	s.log.Info("starting server...")
	if err := s.Svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	<-serverCtx.Done()
}
