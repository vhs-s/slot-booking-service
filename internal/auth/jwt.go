package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 30 * 24 * time.Hour
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	SessionID    string
}

var SecretKeyAccess = []byte("my_secret_key")
var SecretKeyRefresh = []byte("refresh_secret_key")

func GenerateTokenPair(userID string) (*TokenPair, error) {
	sessionID := uuid.NewString()
	accessToken, err := GenerateAccessToken(userID, sessionID)
	if err != nil {
		return nil, err
	}

	claimsRefresh := jwt.MapClaims{
		"jti": uuid.NewString(),
		"sub": userID,
		"sid": sessionID,
		"exp": time.Now().Add(RefreshTokenTTL).Unix(),
	}
	tokenRefresh := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsRefresh)
	refreshToken, err := tokenRefresh.SignedString(SecretKeyRefresh)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionID:    sessionID,
	}, nil
}

func GenerateAccessToken(userID, sessionID string) (string, error) {
	claimsAccess := jwt.MapClaims{
		"user_id": userID,
		"sid":     sessionID,
		"exp":     time.Now().Add(AccessTokenTTL).Unix(),
	}
	tokenAccess := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsAccess)
	return tokenAccess.SignedString(SecretKeyAccess)
}

func ParseToken(tokenString string, secret []byte) (*jwt.Token, error) {
	return jwt.Parse(
		tokenString,
		func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
}

func ExtractClaims(tokenString string, secret []byte) (jwt.MapClaims, bool) {
	token, err := ParseToken(tokenString, secret)
	if err != nil || token == nil || !token.Valid {
		return nil, false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	return claims, ok
}
