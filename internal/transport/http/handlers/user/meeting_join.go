package handlers

import (
	"bytes"
	"errors"
	"html/template"
	"log"
	"net/http"

	bookingdto "slot-booking-service-api/internal/dto/booking"
	"slot-booking-service-api/internal/transport/http/middleware"
	bookingusecase "slot-booking-service-api/internal/usecase/booking"

	"github.com/go-chi/chi/v5"
)

func (uh *UserHandler) MeetingJoinHandler() http.HandlerFunc {
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
		if uh.BookingUC == nil {
			http.Error(w, "Meeting not found", http.StatusNotFound)
			return
		}

		meetingURL, err := uh.BookingUC.JoinMeeting(r.Context(), userID, chi.URLParam(r, "meetingId"))
		if err != nil {
			switch {
			case errors.Is(err, bookingusecase.ErrMeetingNotFound):
				http.Error(w, "Meeting not found", http.StatusNotFound)
			case errors.Is(err, bookingusecase.ErrMeetingFinished):
				uh.renderUnavailableMeeting(w, userID, "Встреча завершена", "Встреча завершена, назначьте новую.")
			case errors.Is(err, bookingusecase.ErrMeetingCancelled):
				uh.renderUnavailableMeeting(w, userID, "Встреча отменена", "Встреча отменена, назначьте новую.")
			default:
				log.Println("Error joining meeting:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		http.Redirect(w, r, meetingURL, http.StatusFound)
	}
}

func (uh *UserHandler) renderUnavailableMeeting(w http.ResponseWriter, userID, title, message string) {
	data := bookingdto.MeetingUnavailableDto{
		Id:            userID,
		Title:         title,
		Message:       message,
		DashboardURL:  "/api/users/" + userID + "/dashboard",
		NewMeetingURL: "/book/" + userID,
	}
	t, err := template.ParseFiles("web/template/header.html", "web/template/meeting_unavailable.html", "web/template/footer.html")
	if err != nil {
		log.Println("Error parse unavailable meeting template:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var page bytes.Buffer
	if err := t.ExecuteTemplate(&page, "meeting_unavailable", data); err != nil {
		log.Println("Error execute unavailable meeting template:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := page.WriteTo(w); err != nil {
		log.Println("Error writing unavailable meeting page:", err)
	}
}
