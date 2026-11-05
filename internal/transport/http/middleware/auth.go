package middleware

import (
	"context"
	"log"
	"net/http"
	"slot-booking-service-api/internal/auth"
)

type contextKey string

const (
	UserKey    contextKey = "userID"
	SessionKey contextKey = "sessionID"
)

func (am *AuthManager) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, sessionID, ok := am.authenticate(w, r)
		if !ok {
			http.Redirect(w, r, "/api/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), UserKey, userID)
		ctx = context.WithValue(ctx, SessionKey, sessionID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuthMiddleware adds the session to the request context when it is
// valid, but still allows guests to access the handler.
func (am *AuthManager) OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, sessionID, ok := am.authenticate(w, r)
		if ok {
			ctx := context.WithValue(r.Context(), UserKey, userID)
			ctx = context.WithValue(ctx, SessionKey, sessionID)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

func (am *AuthManager) authenticate(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	if cookieAccess, err := r.Cookie("access_token"); err == nil {
		if claims, ok := auth.ExtractClaims(cookieAccess.Value, auth.SecretKeyAccess); ok {
			userID, userIDOK := claims["user_id"].(string)
			sessionID, sessionIDOK := claims["sid"].(string)
			if userIDOK && userID != "" && sessionIDOK && sessionID != "" {
				return userID, sessionID, true
			}
		}
	}

	userID, sessionID, ok := am.checkRefreshToken(r)
	if !ok {
		return "", "", false
	}

	accessToken, err := auth.GenerateAccessToken(userID, sessionID)
	if err != nil {
		log.Println("unable to generate access token:", err)
		return "", "", false
	}

	// Keep the refresh token and its server-side session unchanged. Rotating it
	// here makes concurrent requests invalidate each other.
	auth.SetAccessCookie(w, accessToken)
	return userID, sessionID, true
}

func (am *AuthManager) checkRefreshToken(r *http.Request) (string, string, bool) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		return "", "", false
	}

	claims, ok := auth.ExtractClaims(cookie.Value, auth.SecretKeyRefresh)
	if !ok {
		return "", "", false
	}

	userID, userIDOK := claims["sub"].(string)
	sessionID, sessionIDOK := claims["sid"].(string)
	if !userIDOK || userID == "" || !sessionIDOK || sessionID == "" {
		return "", "", false
	}

	cachedToken, err := am.AuthUC.TokenRepository.GetCachedToken(r.Context(), userID, sessionID)
	if err != nil {
		log.Println("unable to validate refresh token:", err)
		return "", "", false
	}
	if cachedToken == "" || cachedToken != cookie.Value {
		return "", "", false
	}

	return userID, sessionID, true
}
