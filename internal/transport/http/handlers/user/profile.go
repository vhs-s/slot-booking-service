package handlers

import (
	"html/template"
	"log"
	"net/http"

	dto "slot-booking-service-api/internal/dto/user"
	"slot-booking-service-api/internal/repositories/postgres/user"
	"slot-booking-service-api/internal/transport/http/middleware"

	"github.com/go-chi/chi/v5"
)

func (uh *UserHandler) ProfileHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID, ok := r.Context().Value(middleware.UserKey).(string)
		if !ok || userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if pathUserID := chi.URLParam(r, "id"); pathUserID != "" && pathUserID != userID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		ur, err := uh.UserUC.UserRepository.GetById(r.Context(), userID)
		if err != nil {
			log.Println("Error get user for profile:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if ur == nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		data := dto.ToGetUserDto(user.ToUserEntity(ur))

		t, err := template.ParseFiles("web/template/header.html", "web/template/profile.html", "web/template/footer.html")
		if err != nil {
			log.Println("Error parse profile template:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := t.ExecuteTemplate(w, "profile", data); err != nil {
			log.Println("Error execute profile template:", err)
		}
	}
}
