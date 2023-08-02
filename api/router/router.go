package router

import (
	"github.com/go-chi/chi"
)

// used to organize all individal handlers.
// all handlers must have the func signature
// func FuncName (w http.ResponseWriter, r *http.Request)
func NewRouter() *chi.Mux {
	r := chi.NewRouter()

	// r.MethodFunc("GET", "/", app.HandleIndex)

	return r
}
