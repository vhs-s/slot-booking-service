package main

import (
	"net/http"
	"os"
	"path/filepath"
	"slot-booking-service-api/internal/database"
	userRepo "slot-booking-service-api/internal/repositories/postgres/user"
	tokenRepo "slot-booking-service-api/internal/repositories/redis/session"
	"slot-booking-service-api/internal/transport/http/handlers"
	systemHandler "slot-booking-service-api/internal/transport/http/handlers/system"
	userHandler "slot-booking-service-api/internal/transport/http/handlers/user"
	"slot-booking-service-api/internal/transport/http/middleware"
	"slot-booking-service-api/internal/usecase/auth"
	userUseCase "slot-booking-service-api/internal/usecase/user"

	"github.com/go-chi/chi/v5"
)

func main() {
	db := database.ConnectPostgres()
	rdb := database.ConnectRedis()

	userRep := userRepo.New(db)
	tokenRep := tokenRepo.New(rdb)

	userUC := userUseCase.New(userRep, tokenRep)
	authUC := auth.New(tokenRep)

	userHand := userHandler.NewUserHandler(userUC, authUC)
	systemHand := systemHandler.NewSystemHandler()

	middlewareManager := middleware.New(authUC)

	router := chi.NewRouter()
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "web", "css", "static"))
	handlers.FileServer(router, "/css/static", filesDir)

	router.Get("/", systemHand.IndexHandler())

	router.Get("/api/register", userHand.Register())
	router.Post("/api/register", userHand.Register())

	router.Get("/api/login", userHand.SignInHandler())
	router.Post("/api/login", userHand.SignInHandler())

	router.Group(func(r chi.Router) {
		r.Use(middlewareManager.AuthMiddleware)

		r.Route("/api/users/{id}", func(r chi.Router) {
			r.Get("/dashboard", userHand.DashBoardHandler())
			r.Get("/schedule", userHand.ScheduleHandler())
			r.Get("/profile", userHand.ProfileHandler())
		})
	})

	http.ListenAndServe(":8080", router)
}
