package server

import (
	"log"
	"net/http"
)

// general error handler.
//
// prints error to console and sends the
// http error response & code.
func ServerErr(w http.ResponseWriter, msg string) {
	log.Printf("[ERROR] %s", msg)
	http.Error(w, msg, http.StatusInternalServerError)
}

func NotFound(w http.ResponseWriter, r *http.Request, msg string) {
	log.Printf("[INFO] resource not found: %s", msg)
	http.NotFound(w, r)
}
