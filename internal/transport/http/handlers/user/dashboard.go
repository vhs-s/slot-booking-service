package handlers

import (
	"html/template"
	"log"
	"net/http"
	"time"

	bookingdto "slot-booking-service-api/internal/dto/booking"
	userdto "slot-booking-service-api/internal/dto/user"
	bookingentity "slot-booking-service-api/internal/entities/booking"
	"slot-booking-service-api/internal/repositories/postgres/user"
	"slot-booking-service-api/internal/transport/http/middleware"

	"github.com/go-chi/chi/v5"
)

func (uh *UserHandler) DashBoardHandler() http.HandlerFunc {
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
			log.Println("Error get user for dashboard:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if ur == nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		udto := userdto.ToGetUserDto(user.ToUserEntity(ur))

		bookings := make([]bookingentity.Booking, 0)
		if uh.BookingUC != nil {
			bookings, err = uh.BookingUC.ListByUserId(r.Context(), userID)
			if err != nil {
				log.Println("Error get bookings for dashboard:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		}

		location := time.UTC
		if uh.ScheduleUC != nil {
			userCalendar, _, err := uh.ScheduleUC.Get(r.Context(), userID)
			if err != nil {
				log.Println("Error get calendar timezone for dashboard:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if userCalendar != nil {
				location, err = time.LoadLocation(userCalendar.Timezone)
				if err != nil {
					log.Println("Error load calendar timezone for dashboard:", err)
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
			}
		}

		filter := bookingdto.ParseMeetingStatusFilter(r.URL.Query().Get("status"))
		data := bookingdto.NewDashboardDto(udto, bookings, location, filter, time.Now())
		t, err := template.ParseFiles("web/template/header.html", "web/template/dashboard.html", "web/template/footer.html")
		if err != nil {
			log.Println("Error parse dashboard template:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := t.ExecuteTemplate(w, "dashboard", data); err != nil {
			log.Println("Error execute dashboard template:", err)
		}
	}
}
