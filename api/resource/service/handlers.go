package service

import (
	"net/http"

	"github.com/nimbus/pkg/server"
)

func (s *server.Server) HandlerExample() http.HandlerFunc {
	return func(http.ResponseWriter, *http.Request) {

	}
}
