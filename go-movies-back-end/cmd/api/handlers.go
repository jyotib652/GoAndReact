package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v4"
)

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprintf(w, "Hello world from %s", app.Domain)
	var payload = struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Version string `json:"version"`
	}{
		Status:  "active",
		Message: "Go movies up and running",
		Version: "1.0.0",
	}

	// // we use json.Marshal to convert other data as JSON data
	// out, err := json.Marshal(payload)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// // to send JSON data, we have to declare that we are sending JSON data
	// // and to do that we must set the header
	// w.Header().Set("Content-type", "application/json")
	// // appending to the header that everything worked out as it should be(Sending OK. http 200)
	// w.WriteHeader(http.StatusOK)
	// w.Write(out)

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) AllMovies(w http.ResponseWriter, r *http.Request) {
	// var movies []models.Movie

	// // when we want to parse a string as time. we have to use time.Parse
	// // "2006-01-01" is the layout for time string where we are mentioning to the time.Parse
	// // that expect first 4 digit character(2006) as year(1983) then a dash(-) then two digit
	// // character(01) as month(03) then a dash(-) then again two digit character(01) as day(07)
	// rd, _ := time.Parse("2006-01-01", "1986-03-07")

	// highlander := models.Movie{
	// 	ID:          1,
	// 	Title:       "Highlander",
	// 	ReleaseDate: rd,
	// 	MPAARating:  "R",
	// 	RunTime:     116, // in minutes. i.e. 1 hour 56 minutes
	// 	Description: "A very nice movie.",
	// 	CreatedAt:   time.Now(),
	// 	UpdatedAt:   time.Now(),
	// }

	// movies = append(movies, highlander)

	// rd, _ = time.Parse("2006-01-01", "1981-06-12")

	// rotla := models.Movie{
	// 	ID:          2,
	// 	Title:       "Raiders of the Lost Arc",
	// 	ReleaseDate: rd,
	// 	MPAARating:  "PG-13",
	// 	RunTime:     115, // in minutes. i.e. 1 hour 55 minutes
	// 	Description: "Another very nice movie.",
	// 	CreatedAt:   time.Now(),
	// 	UpdatedAt:   time.Now(),
	// }

	// movies = append(movies, rotla)

	movies, err := app.DB.AllMovies()
	if err != nil {
		// fmt.Println(err)
		app.errorJSON(w, err)
		return
	}

	// // we use json.Marshal to convert other data as JSON data
	// out, err := json.Marshal(movies)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// // to send JSON data, we have to declare that we are sending JSON data
	// // and to do that we must set the header
	// w.Header().Set("Content-type", "application/json")
	// // appending to the header that everything worked out as it should be(Sending OK. http 200)
	// w.WriteHeader(http.StatusOK)
	// w.Write(out)

	_ = app.writeJSON(w, http.StatusOK, movies)
}

func (app *application) authenticate(w http.ResponseWriter, r *http.Request) {
	// read json payload
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// validate user against database
	user, err := app.DB.GetUserByEmail(requestPayload.Email)
	if err != nil {
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	// check password
	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	// create a jwt user
	u := jwtUser{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	// generate tokens
	tokens, err := app.auth.GenerateTokenPair(&u)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// log.Println(tokens.Token)
	// set the refresh cookie. set the refresh cookie to the client's browser
	refreshCookie := app.auth.GetRefreshCookie(tokens.RefreshToken)
	http.SetCookie(w, refreshCookie)

	// sending the tokens to the client. Writing the token string to the browser
	// w.Write([]byte(tokens.Token))

	app.writeJSON(w, http.StatusAccepted, tokens)
}

func (app *application) refreshToken(w http.ResponseWriter, r *http.Request) {
	// find the cookie with refreshToken among all the cookies that came with user request
	for _, cookie := range r.Cookies() {
		if cookie.Name == app.auth.CookieName {
			claims := &Claims{}
			refreshToken := cookie.Value

			// parse the refresh token to get the claims. Claims are several information regarding the user
			_, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(app.JWTSecret), nil
			})
			if err != nil {
				app.errorJSON(w, errors.New("unauthorized"), http.StatusUnauthorized)
				return
			}

			// get the user id from the refresh token claims
			userID, err := strconv.Atoi(claims.Subject)
			if err != nil {
				app.errorJSON(w, errors.New("unknown user"), http.StatusUnauthorized)
				return
			}

			// get the user by id(the user id from claims) from the database
			user, err := app.DB.GetUserByID(userID)
			if err != nil {
				app.errorJSON(w, errors.New("unknown user"), http.StatusUnauthorized)
				return
			}

			// create new jwtuser
			u := jwtUser{
				ID:        user.ID,
				FirstName: user.FirstName,
				LastName:  user.LastName,
			}

			// now generate token pairs
			tokenPairs, err := app.auth.GenerateTokenPair(&u)
			if err != nil {
				app.errorJSON(w, errors.New("error generating tokens"), http.StatusUnauthorized)
				return
			}

			// send a new refresh token cookie as response
			http.SetCookie(w, app.auth.GetRefreshCookie(tokenPairs.RefreshToken))

			// sending the token pairs as JSON also
			app.writeJSON(w, http.StatusOK, tokenPairs)
		}
	}

}

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, app.auth.GetExpiredRefreshCookie())
	w.WriteHeader(http.StatusAccepted)
}

func (app *application) MovieCatalog(w http.ResponseWriter, r *http.Request) {
	movies, err := app.DB.AllMovies()
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, movies)
}
