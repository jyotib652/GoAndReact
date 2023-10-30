package models

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"-"` // time.Time for timestamp fields
	UpdatedAt time.Time `json:"-"` // time.Time for timestamp fields
}

// this function takes a plain text password and match it with the hash of a password stored in the database
func (u *User) PasswordMatches(plainText string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plainText))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			// Invalid password
			return false, nil
		default:
			// Something else has happened that caused the error
			return false, err
		}
	}

	return true, nil
}
