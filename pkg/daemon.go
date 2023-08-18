package pkg

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

// start sfs lister daemon
func daemon() {
	log.Printf("[DEBUG] starting sfs daemon...")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// start http listener server. should be able to run in
	// the background without user interaction.

	// go startHTTPServer()

	<-stop // Block until an interrupt signal is received
	log.Printf("[DEBUG] received interrupt signal. stopping daemon...")
}
