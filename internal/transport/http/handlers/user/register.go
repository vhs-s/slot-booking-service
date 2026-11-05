package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"slot-booking-service-api/internal/auth"
	userE "slot-booking-service-api/internal/entities/user"
	userR "slot-booking-service-api/internal/repositories/postgres/user"
)

func (uh *UserHandler) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//check user in database
		ctx := r.Context()
		switch r.Method {
		case http.MethodGet:
			t, err := template.ParseFiles("web/template/header.html", "web/template/register.html", "web/template/footer.html")
			if err != nil {
				log.Println("Error parse template:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if err := t.ExecuteTemplate(w, "register", nil); err != nil {
				log.Println("Error execute template:", err)
			}
		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Invalid form", http.StatusBadRequest)
				return
			}
			name := r.Form.Get("name")
			email := r.Form.Get("email")
			password := r.Form.Get("password")
			if name == "" || email == "" || password == "" {
				http.Error(w, "Name, email and password are required", http.StatusBadRequest)
				return
			}

			ue := userE.New(name, email, password)
			ur := userR.ToUserRecord(ue)
			if err := uh.UserUC.UserRepository.Create(ctx, ur); err != nil {
				log.Println("Error create user:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			tokenPair, err := auth.GenerateTokenPair(ue.Id)
			if err != nil {
				log.Println("GenerateTokenPair error:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if err := uh.AuthUC.TokenRepository.SetCachedToken(ctx, ue.Id, tokenPair.SessionID, tokenPair.RefreshToken); err != nil {
				log.Println("SetCachedToken error:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			auth.SetAccessCookie(w, tokenPair.AccessToken)
			auth.SetRefreshCookie(w, tokenPair.RefreshToken)

			redirectUrl := fmt.Sprintf("/api/users/%s/dashboard", ue.Id)
			http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
		default:
			w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
