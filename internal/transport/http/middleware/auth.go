package middleware

import (
	"context"
	"net/http"
	"slot-booking-service-api/internal/auth"
)

type contextKey string

const UserKey contextKey = "userID"

func (am *AuthManager) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string

		cookieAccess, err := r.Cookie("access_token")
		if err == nil {
			tokenString = cookieAccess.Value
		}

		token, err := auth.ParseToken(tokenString, auth.SecretKeyAccess)

		if err != nil || token == nil || !token.Valid {
			valid, userId := am.checkRefreshToken(r)
			if !valid {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			tp, err := auth.GenerateTokenPair(userId)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			err = am.AuthUC.TokenRepository.SetCachedToken(r.Context(), userId, tp.RefreshToken)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			auth.SetAccessCookie(w, tp.AccessToken)
			auth.SetRefreshCookie(w, tp.RefreshToken)

			tokenString = tp.AccessToken
			auth.SetAccessCookie(w, tp.AccessToken)
			auth.SetRefreshCookie(w, tp.RefreshToken)
		}

		claims, ok := auth.ExtractClaims(tokenString, auth.SecretKeyAccess)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		userId, userIdOk := claims["user_id"].(string)
		if !userIdOk || userId == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), UserKey, userId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (am *AuthManager) checkRefreshToken(r *http.Request) (bool, string) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		return false, ""
	}

	tokenString := cookie.Value
	token, err := auth.ParseToken(tokenString, auth.SecretKeyRefresh)
	if err != nil || token == nil || !token.Valid {
		return false, ""
	}

	claims, ok := auth.ExtractClaims(tokenString, auth.SecretKeyRefresh)
	if !ok {
		return false, ""
	}

	userIdCookie, exists := claims["sub"]
	if !exists || userIdCookie == nil {
		return false, ""
	}

	userIdStr, ok := userIdCookie.(string)
	if !ok {
		return false, ""
	}

	cachedToken, err := am.AuthUC.TokenRepository.GetCachedToken(context.Background(), userIdStr)
	if err != nil || cachedToken != tokenString {
		return false, ""
	}

	return true, userIdStr
}
