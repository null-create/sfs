package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

/*
TODO:
	design API specifications

	New to come up with repsectieve GET, POST, PUT, DELETE, etc.
	methods and /nimbus/ (or /n/usr/root ?) etc routes, and
	associated handlers
*/

// based off example from: https://github.com/go-chi/chi/blob/master/README.md
func NewRouter() *chi.Mux {

	// instantiate router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})

	// // RESTy example routes for "articles" resource
	// r.Route("/articles", func(r chi.Router) {
	// 	r.With(paginate).Get("/", listArticles)                           // GET /articles
	// 	r.With(paginate).Get("/{month}-{day}-{year}", listArticlesByDate) // GET /articles/01-16-2017

	// 	r.Post("/", createArticle)       // POST /articles
	// 	r.Get("/search", searchArticles) // GET /articles/search

	// 	// Regexp url parameters:
	// 	r.Get("/{articleSlug:[a-z-]+}", getArticleBySlug) // GET /articles/home-is-toronto

	// 	// Subrouters:
	// 	r.Route("/{articleID}", func(r chi.Router) {
	// 		r.Use(ArticleCtx)
	// 		r.Get("/", getArticle)       // GET /articles/123
	// 		r.Put("/", updateArticle)    // PUT /articles/123
	// 		r.Delete("/", deleteArticle) // DELETE /articles/123
	// 	})
	// })

	// Mount the admin sub-router
	r.Mount("/admin", adminRouter())

	return r
}

// A completely separate router for administrator routes
func adminRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(AdminOnly)
	// r.Get("/", adminIndex)
	// r.Get("/accounts", adminListAccounts)
	return r
}

func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, ok := ctx.Value("acl.permission").(float64)
		if !ok {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
