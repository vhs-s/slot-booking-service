package handlers

import (
	"log"
	"net/http"
	"slot-booking-service-api/internal/auth"
	"slot-booking-service-api/internal/transport/http/middleware"
)

func (uh *UserHandler) LogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID, userIDOK := r.Context().Value(middleware.UserKey).(string)
		sessionID, sessionIDOK := r.Context().Value(middleware.SessionKey).(string)
		if !userIDOK || userID == "" || !sessionIDOK || sessionID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if err := uh.AuthUC.TokenRepository.DeleteCachedToken(r.Context(), userID, sessionID); err != nil {
			log.Println("Error delete user session:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		auth.ClearAuthCookies(w)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
