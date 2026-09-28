package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"text/template"

	"slot-booking-service-api/internal/auth"
	"slot-booking-service-api/internal/repositories/postgres/user"
)

func (uh *UserHandler) SignInHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			t, err := template.ParseFiles(
				"web/template/header.html",
				"web/template/login.html",
				"web/template/footer.html",
			)
			if err != nil {
				log.Println("Error parse template:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if err := t.ExecuteTemplate(w, "login", nil); err != nil {
				log.Println("Error execute template:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

		case http.MethodPost:
			email := r.FormValue("email")
			password := r.FormValue("password")

			if email == "" || password == "" {
				log.Println("Login failed: empty email or password")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			log.Println("email from form:", email)

			ur, err := uh.UserUC.UserRepository.GetByEmail(r.Context(), email)

			log.Println("GetByEmail result:", ur)
			log.Println("GetByEmail error:", err)

			if err != nil {
				log.Println("Login failed: user not found:", err)
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			ue := user.ToUserEntity(ur)

			h := sha256.New()
			h.Write([]byte(password))
			hashedData := h.Sum(nil)
			hexHash := hex.EncodeToString(hashedData)

			if hexHash != ue.PasswordHash {
				log.Println("Login failed: invalid password")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			tokenPair, err := auth.GenerateTokenPair(ue.Id)
			if err != nil {
				log.Println("GenerateTokenPair error:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			err = uh.AuthUC.TokenRepository.SetCachedToken(r.Context(), ue.Id, tokenPair.RefreshToken)
			if err != nil {
				log.Println("SetCachedToken error:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			auth.SetAccessCookie(w, tokenPair.AccessToken)
			auth.SetRefreshCookie(w, tokenPair.RefreshToken)

			log.Println("Login successful:", ue.Id)

			http.Redirect(w, r, "/", http.StatusSeeOther)
			return

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}
