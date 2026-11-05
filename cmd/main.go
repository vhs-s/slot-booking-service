package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	jwtAuth "slot-booking-service-api/internal/auth"
	"slot-booking-service-api/internal/config"
	"slot-booking-service-api/internal/database"
	bookingRepo "slot-booking-service-api/internal/repositories/postgres/booking"
	scheduleRepo "slot-booking-service-api/internal/repositories/postgres/schedule"
	userRepo "slot-booking-service-api/internal/repositories/postgres/user"
	tokenRepo "slot-booking-service-api/internal/repositories/redis/session"
	"slot-booking-service-api/internal/transport/http/handlers"
	bookingHandler "slot-booking-service-api/internal/transport/http/handlers/booking"
	systemHandler "slot-booking-service-api/internal/transport/http/handlers/system"
	userHandler "slot-booking-service-api/internal/transport/http/handlers/user"
	"slot-booking-service-api/internal/transport/http/middleware"
	"slot-booking-service-api/internal/usecase/auth"
	bookingUseCase "slot-booking-service-api/internal/usecase/booking"
	scheduleUseCase "slot-booking-service-api/internal/usecase/schedule"
	userUseCase "slot-booking-service-api/internal/usecase/user"

	"github.com/go-chi/chi/v5"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки конфигурации:", err)
	}

	jwtAuth.SecretKeyAccess = []byte(cfg.JWTAccessSecret)
	jwtAuth.SecretKeyRefresh = []byte(cfg.JWTRefreshSecret)

	db := database.ConnectPostgres(cfg.PostgresDSN)
	rdb := database.ConnectRedis(cfg.RedisDSN)

	userRep := userRepo.New(db)
	scheduleRep := scheduleRepo.New(db)
	bookingRep := bookingRepo.New(db)
	tokenRep := tokenRepo.New(rdb)

	userUC := userUseCase.New(userRep, tokenRep)
	scheduleUC := scheduleUseCase.New(scheduleRep)
	bookingUC := bookingUseCase.New(bookingRep, scheduleUC)
	authUC := auth.New(tokenRep)

	userHand := userHandler.NewUserHandler(userUC, authUC, scheduleUC)
	userHand.BookingUC = bookingUC
	bookingHand := bookingHandler.NewBookingHandler(userUC, bookingUC)
	systemHand := systemHandler.NewSystemHandler(userUC)

	middlewareManager := middleware.New(authUC)

	router := chi.NewRouter()
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "web", "css", "static"))
	handlers.FileServer(router, "/css/static", filesDir)

	router.With(middlewareManager.OptionalAuthMiddleware).Get("/", systemHand.IndexHandler())

	router.Get("/api/register", userHand.Register())
	router.Post("/api/register", userHand.Register())

	router.Get("/api/login", userHand.SignInHandler())
	router.Post("/api/login", userHand.SignInHandler())

	router.Group(func(r chi.Router) {
		r.Use(middlewareManager.OptionalAuthMiddleware)

		r.Get("/book/{id}", bookingHand.CalendarHandler())
		r.Post("/book/{id}", bookingHand.CalendarHandler())
	})

	router.Group(func(r chi.Router) {
		r.Use(middlewareManager.AuthMiddleware)

		r.Route("/api/users/{id}", func(r chi.Router) {
			r.Get("/dashboard", userHand.DashBoardHandler())
			r.Get("/meetings/{meetingId}/join", userHand.MeetingJoinHandler())
			r.Get("/schedule", userHand.ScheduleHandler())
			r.Post("/schedule", userHand.ScheduleHandler())
			r.Get("/profile", userHand.ProfileHandler())
			r.Get("/logout", userHand.LogoutHandler())
			r.Post("/logout", userHand.LogoutHandler())
		})
	})

	if err := http.ListenAndServe(":"+cfg.HTTPPort, router); err != nil {
		log.Fatal("Ошибка запуска HTTP-сервера:", err)
	}
}
