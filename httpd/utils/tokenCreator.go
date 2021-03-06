package utils

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// HmacSampleSecret this is the random string needed to sign/encrypt/decrypt the
// JWT tokens
var HmacSampleSecret []byte

// TokenCreator creates a jwt token for us to use
func TokenCreator(email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":     email,
		"expiresAt": time.Now().Add(time.Hour * 100).Unix(), // tokens are valid for 10 minutes?
	})
	tokenString, err := token.SignedString(HmacSampleSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
