package auth

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

var SecretKeyAccess = []byte("my_secret_key")
var SecretKeyRefresh = []byte("refresh_secret_key")

func GenerateTokenPair(userID string) (*TokenPair, error) {
	claimsAccess := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Minute * 15).Unix(),
	}

	tokenAccess := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsAccess)

	claimsRefresh := jwt.MapClaims{
		"jti": uuid.NewString(),
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	}

	tokenRefresh := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsRefresh)

	tokenAccessStr, err := tokenAccess.SignedString(SecretKeyAccess)

	if err != nil {
		return nil, err
	}

	tokenRefreshStr, err := tokenRefresh.SignedString(SecretKeyRefresh)

	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  tokenAccessStr,
		RefreshToken: tokenRefreshStr,
	}, nil
}

func ParseToken(tokenString string, secret []byte) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
}

func ExtractClaims(tokenString string, secret []byte) (jwt.MapClaims, bool) {
	hmacSecretString := secret
	hmacSecret := []byte(hmacSecretString)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return hmacSecret, nil
	})

	if err != nil {
		return nil, false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, true
	} else {
		log.Printf("Invalid JWT Token")
		return nil, false
	}
}
