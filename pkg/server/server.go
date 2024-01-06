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
// contained within the router
func NewServer() *Server {
	// router contains (and intializes) the server-side SNS service instance.
	rtr := NewRouter()
	cfg := ServerConfig()

	return &Server{
		Svr: &http.Server{
			Handler:      rtr,
			Addr:         cfg.Addr,
			ReadTimeout:  cfg.TimeoutRead,
			WriteTimeout: cfg.TimeoutWrite,
			IdleTimeout:  cfg.TimeoutIdle,
		},
	}
}

func secondsToTimeStr(seconds float64) string {
	duration := time.Duration(int64(seconds)) * time.Second
	timeValue := time.Time{}.Add(duration)
	return timeValue.Format("15:04:05")
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
				log.Print("[WARNING] shutdown timed out. forcing exit.")
				if _, err := s.Shutdown(); err != nil {
					log.Fatal(err)
				}
			}
		}()

		log.Printf("[INFO] shutting down server...")
		if err := s.Svr.Shutdown(shutdownCtx); err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	log.Printf("[INFO] starting server...")
	if err := s.Svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	<-serverCtx.Done()
}

// start a server that can be shut down using a shutDown bool channel.
func (s *Server) Start(shutDown chan bool) {
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	go func() {
		// blocks until turnOff = true
		// (set by outer test and passed after checks are completed (or failed))
		<-shutDown

		// shutdown signal with grace period of 10 seconds
		shutdownCtx, _ := context.WithTimeout(serverCtx, 10*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Print("[WARNING] shutdown timed out. forcing exit.")
				if _, err := s.Shutdown(); err != nil {
					log.Fatal(err)
				}
			}
		}()

		log.Printf("[INFO] shutting down server...")
		err := s.Svr.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	log.Printf("[INFO] starting server...")
	if err := s.Svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	<-serverCtx.Done()
}
