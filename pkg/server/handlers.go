package server

import "net/http"

func (s *Server) HandlerExample() http.HandlerFunc {
	// any pre-op handling can go here

	return func(http.ResponseWriter, *http.Request) {

	}
}
