package main

import "net/http"

func (app *application) enableCORS(h http.Handler) http.Handler {
	// Here, we are simply just modifying the request as it comes in.
	// Note: midllewares acts on requests. Before going to the handlers
	// middlewares act on the request. After middlewares finished with
	// acting on the request, requests goes to handlers.

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://*")

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, X-CSRF-Token, Authorization")
			return
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

func (app *application) authRequired(next http.Handler) http.Handler {
	// Since we need to access both the responsewriter and request so we
	// would do the same things that we diid in enableCORS() function/method
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// we don't care about the token and claims that's why we are throwing them out and only storing the error
		_, _, err := app.auth.GetTokenFromHeaderAndVerify(w, r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
