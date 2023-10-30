package main

import (
	"backend/internal/repository"
	"backend/internal/repository/dbrepo"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

const port = 8080

type application struct {
	DSN    string // DSN = Data Source Name
	Domain string
	// DB     *sql.DB
	DB           repository.DatabaseRepo
	auth         Auth
	JWTSecret    string
	JWTIssuer    string
	JWTAudience  string
	CookieDomain string
}

func main() {
	// set application config
	var app application

	// read from command line (flags)
	flag.StringVar(&app.DSN, "dsn", "host=localhost port=5432 user=postgres password=postgres dbname=movies sslmode=disable timezone=UTC connect_timeout=5", "Postgres connection string")
	flag.StringVar(&app.JWTSecret, "jwt-seret", "verysecret", "signing secret")
	flag.StringVar(&app.JWTIssuer, "jwt-issuer", "example.com", "signing issuer")
	flag.StringVar(&app.JWTAudience, "jwt-audience", "example.com", "signing audience")
	flag.StringVar(&app.CookieDomain, "cookie-domain", "localhost", "cookie domain")
	flag.StringVar(&app.Domain, "domain", "example.com", "domain")
	flag.Parse()

	// connect to the database
	conn, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}
	// app.DB = conn
	app.DB = &dbrepo.PostgresDBRepo{DB: conn}
	// defer app.DB.Close()
	// defer conn.Close()
	defer app.DB.Connection().Close()

	app.auth = Auth{
		Issuer:        app.JWTIssuer,
		Audience:      app.JWTAudience,
		Secret:        app.JWTSecret,
		TokenExpiry:   time.Minute * 15,
		RefreshExpiry: time.Hour * 24,
		CookiePath:    "/", // "/" means root level of our app which means cookie is good for anywhere in our app.
		CookieName:    "__Host-refresh_token",
		CookieDomain:  app.CookieDomain,
	}

	// app.Domain = "example.com"

	log.Println("Starting application on port", port)

	// http.HandleFunc("/", Hello) // from now, we'll be using app.routes()
	// start a webserver
	// ListenAndServe needs a port and a handler. We are using nil for handler for now.
	// Since, we've created we don't need to pass nil, we'll use app.routes() handler
	// err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), app.routes())
	if err != nil {
		log.Fatal(err)
	}
}
