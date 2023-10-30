package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	// create a router mux. Mux stands for multiplexor
	mux := chi.NewRouter()

	// use a middleware. We are using a middleware which comes from chi package.
	// middleware.Recoverer-> what it does? when your application panics, it will
	// log it along with back trace; It will send appropriate http header which is http 500
	// there was some kind of internal server error. And it will bring things back up so that
	// your application does not grind to a halt.

	mux.Use(middleware.Recoverer)
	// apply CORS
	mux.Use(app.enableCORS) // our custom middleware applies to all the following routes

	mux.Get("/", app.Home)

	mux.Post("/authenticate", app.authenticate)
	mux.Get("/refresh", app.refreshToken)
	mux.Get("/logout", app.logout)

	mux.Get("/movies", app.AllMovies)

	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(app.authRequired) // authRequired middleware only applies to the following routes in this block

		mux.Get("/movies", app.MovieCatalog) // actual route is "/admin/movies"
	})

	return mux
}
