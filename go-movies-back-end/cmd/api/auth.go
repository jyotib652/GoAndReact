package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Auth struct {
	Issuer        string
	Audience      string
	Secret        string
	TokenExpiry   time.Duration
	RefreshExpiry time.Duration
	CookieDomain  string // what is the domain associated with cookie. Something like example.com
	CookiePath    string
	CookieName    string
}

type jwtUser struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type TokenPairs struct {
	Token        string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Claims struct {
	jwt.RegisteredClaims
}

func (j *Auth) GenerateTokenPair(user *jwtUser) (TokenPairs, error) {
	// Create a token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set the claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	// "sub", "aud", "iss", "iat", "typ" : these are must and shouldn't be renamed at all
	claims["sub"] = fmt.Sprint(user.ID)     // sub for subject. fmt.Sprint(user.ID) => here, we're converting user.ID as string
	claims["aud"] = j.Audience              // aud for audience
	claims["iss"] = j.Issuer                // iss for issuer
	claims["iat"] = time.Now().UTC().Unix() // iat for issued at
	claims["typ"] = "JWT"                   // typ for type. For type of token we are generating.

	// Set the expiry for JWT
	claims["exp"] = time.Now().Add(j.TokenExpiry).Unix() // exp for expiry.

	// Create a signed token
	signedAccessToken, err := token.SignedString([]byte(j.Secret))
	if err != nil {
		return TokenPairs{}, err
	}

	// Create a refresh token and set claims
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshTokenClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshTokenClaims["sub"] = fmt.Sprint(user.ID)
	refreshTokenClaims["iat"] = time.Now().UTC().Unix()

	// Set the expiry for refresh token
	refreshTokenClaims["exp"] = time.Now().UTC().Add(j.RefreshExpiry).Unix()

	// Create signed refresh token
	signedRefreshToken, err := refreshToken.SignedString([]byte(j.Secret))
	if err != nil {
		return TokenPairs{}, err
	}

	// Create TokenPairs(struct) and populate with signed tokens
	var tokenPairs = TokenPairs{
		Token:        signedAccessToken,
		RefreshToken: signedRefreshToken,
	}

	// Return TokenPairs
	return tokenPairs, nil
}

func (j *Auth) GetRefreshCookie(refreshToken string) *http.Cookie {
	return &http.Cookie{
		Name:     j.CookieName,
		Path:     j.CookiePath,
		Value:    refreshToken,
		Expires:  time.Now().Add(j.RefreshExpiry),
		MaxAge:   int(j.RefreshExpiry.Seconds()),
		SameSite: http.SameSiteStrictMode, // make this cookie limited only to this site
		Domain:   j.CookieDomain,
		HttpOnly: true, // making the cookie secure. So that Javascript doesn't have access to this cookie at all in a Web Browser.
		Secure:   true, // making the cookie secure. Although in development it won't be secure but in production environment, it will be secure
	}
}

// to delete the cookie from the browser
func (j *Auth) GetExpiredRefreshCookie() *http.Cookie {
	return &http.Cookie{
		Name:     j.CookieName,
		Path:     j.CookiePath,
		Value:    "",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		SameSite: http.SameSiteStrictMode, // make this cookie limited only to this site
		Domain:   j.CookieDomain,
		HttpOnly: true, // making the cookie secure. So that Javascript doesn't have access to this cookie at all in a Web Browser.
		Secure:   true, // making the cookie secure. Although in development it won't be secure but in production environment, it will be secure
	}
}

// protecting routes with jwt tokens -- to do that we are mentioning in our request header (authorization header)
// bearer token. So when users' request hit any route we would extract the bearer token and validate the token.
// if it's valid then only we allow user to access that resource.
func (j *Auth) GetTokenFromHeaderAndVerify(w http.ResponseWriter, r *http.Request) (string, *Claims, error) {
	// add a header to our response (this is a good practice)
	w.Header().Add("Vary", "Authorization")

	// go to the request and check for the auth header
	authHeader := r.Header.Get("Authorization")

	// sanity check (if authorization header exists or not)
	if authHeader == "" {
		return "", nil, errors.New("no auth header")
	}

	// split the header on spaces [Two parts:- "Bearer Token"]
	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 {
		return "", nil, errors.New("invalid auth header")
	}

	// check to see if we have the word Bearer
	if headerParts[0] != "Bearer" {
		return "", nil, errors.New("invalid auth header")
	}

	// actual Bearer token
	token := headerParts[1]

	// declare an empty claims
	claims := &Claims{}

	// parse token and read the values and store them into claims
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
		}
		return []byte(j.Secret), nil
	})

	// this error may also happen due to the expiry of a token [for an expired token]
	if err != nil {
		if strings.HasPrefix(err.Error(), "token is expired by") {
			return "", nil, errors.New("expired token")
		}

		return "", nil, err
	}

	// check if we issued this token or not
	if claims.Issuer != j.Issuer {
		return "", nil, errors.New("invalid issuer")
	}

	// if we get pass the above logics then we have a valid non expired jwt token.
	// So we should return the token, claims and no error now.
	return token, claims, nil

}
