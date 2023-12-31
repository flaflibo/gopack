package utils

import (
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

func VerifyJwt(token string, secret string) (claims jwt.MapClaims, err error) {
	// Parse the token.
	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Return the secret.
		return []byte(secret), nil
	})

	// Check for errors.
	if err != nil {
		return nil, err
	}

	// If the token is valid, return the claims.
	if tokenParsed.Valid {
		return tokenParsed.Claims.(jwt.MapClaims), nil
	}

	// Otherwise, return an error.
	return nil, fmt.Errorf("invalid token")
}

func CreateToken(secret string, claims interface{}) (string, error) {
	signingKey := []byte(secret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims.(jwt.Claims))

	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}

	return tokenString, err
}
