package handlers

import (
	"fmt"
	"log"
	"net/http"
	"slot-booking-service-api/internal/auth"
	dto "slot-booking-service-api/internal/dto/user"
	"slot-booking-service-api/internal/repositories/postgres/user"
	"text/template"
)

func (uh *UserHandler) DashBoardHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			t, err := template.ParseFiles("web/template/header.html", "web/template/dashboard.html", "web/template/footer.html")
			if err != nil {
				log.Println("Error parse template", err)
			}

			cookieAccess, err := r.Cookie("access_token")
			if err != nil {
				http.Redirect(w, r, "/api/login", http.StatusSeeOther)
			}
			tokenString := cookieAccess.Value
			fmt.Println("TOKEN STRING:", tokenString)
			claims, _ := auth.ExtractClaims(tokenString, auth.SecretKeyAccess)
			userId, _ := claims["user_id"].(string)
			fmt.Println("USER ID:", userId)
			ur, _ := uh.UserUC.UserRepository.GetById(r.Context(), userId)
			ue := user.ToUserEntity(ur)
			udto := dto.ToGetUserDto(ue)

			t.ExecuteTemplate(w, "dashboard", udto)
			break
		case http.MethodPost:
		}
	}
}
