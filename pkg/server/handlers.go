package server

import "net/http"

func (s *Server) HandlerExample() http.HandlerFunc {
	return func(http.ResponseWriter, *http.Request) {

	}
}
