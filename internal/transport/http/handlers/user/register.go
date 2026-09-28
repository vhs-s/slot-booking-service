package handlers

import (
	"fmt"
	"log"
	"net/http"
	"slot-booking-service-api/internal/auth"
	userE "slot-booking-service-api/internal/entities/user"
	userR "slot-booking-service-api/internal/repositories/postgres/user"
	"text/template"
)

func (uh *UserHandler) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//check user in database
		ctx := r.Context()
		switch r.Method {
		case http.MethodGet:
			t, err := template.ParseFiles("web/template/header.html", "web/template/register.html", "web/template/footer.html")
			if err != nil {
				log.Println("Error parse template", err)
			}
			t.ExecuteTemplate(w, "register", nil)
			break
		case http.MethodPost:
			r.ParseForm()
			name := r.Form.Get("name")
			email := r.Form.Get("email")
			password := r.Form.Get("password")

			ue := userE.New(name, email, password)
			ur := userR.ToUserRecord(ue)
			tokenPair, _ := auth.GenerateTokenPair(ue.Id)

			auth.SetAccessCookie(w, tokenPair.AccessToken)
			auth.SetRefreshCookie(w, tokenPair.RefreshToken)

			uh.AuthUC.TokenRepository.SetCachedToken(ctx, ue.Id, tokenPair.RefreshToken)
			uh.UserUC.UserRepository.Create(ctx, ur)

			redirectUrl := fmt.Sprintf("/api/users/%s/dashboard", ue.Id)
			http.Redirect(w, r, redirectUrl, 303)
		}
	}
}
