package handlers

import (
	"html/template"
	"log"
	"net/http"
	dto "slot-booking-service-api/internal/dto/user"
	"slot-booking-service-api/internal/repositories/postgres/user"
	"slot-booking-service-api/internal/transport/http/middleware"
)

func (sh *SystemHandler) IndexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles("web/template/header.html", "web/template/index.html", "web/template/footer.html")
		if err != nil {
			log.Println("Error parse template", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		var data *dto.GetUserDto
		if userID, ok := r.Context().Value(middleware.UserKey).(string); ok && userID != "" {
			ur, err := sh.UserUC.UserRepository.GetById(r.Context(), userID)
			if err != nil {
				log.Println("Error get user for index:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if ur != nil {
				data = dto.ToGetUserDto(user.ToUserEntity(ur))
			}
		}

		if err := t.ExecuteTemplate(w, "index", data); err != nil {
			log.Println("Error execute template:", err)
		}
	}
}
