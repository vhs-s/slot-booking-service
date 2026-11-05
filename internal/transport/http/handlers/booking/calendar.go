package handlers

import (
	"bytes"
	"errors"
	"html/template"
	"log"
	"net/http"
	"time"

	bookingdto "slot-booking-service-api/internal/dto/booking"
	"slot-booking-service-api/internal/transport/http/middleware"
	bookingusecase "slot-booking-service-api/internal/usecase/booking"

	"github.com/go-chi/chi/v5"
)

func (bh *BookingHandler) CalendarHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			bh.renderCalendar(w, r, r.URL.Query().Get("date"), bookingusecase.CreateInput{}, "", http.StatusOK)
		case http.MethodPost:
			bh.createBooking(w, r)
		default:
			w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (bh *BookingHandler) createBooking(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	userId := chi.URLParam(r, "id")
	input := bookingusecase.CreateInput{
		Date:       r.PostForm.Get("date"),
		Time:       r.PostForm.Get("time"),
		GuestName:  r.PostForm.Get("guest_name"),
		GuestEmail: r.PostForm.Get("guest_email"),
		Topic:      r.PostForm.Get("topic"),
		Comment:    r.PostForm.Get("comment"),
		MeetingURL: r.PostForm.Get("meeting_url"),
	}

	host, err := bh.hostName(r, userId)
	if err != nil {
		if errors.Is(err, bookingusecase.ErrCalendarNotFound) {
			http.Error(w, "Calendar not found", http.StatusNotFound)
			return
		}
		log.Println("Error getting calendar owner:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	created, err := bh.BookingUC.Create(r.Context(), userId, input)
	if err != nil {
		switch {
		case errors.Is(err, bookingusecase.ErrCalendarNotFound):
			http.Error(w, "Calendar not found", http.StatusNotFound)
		case errors.Is(err, bookingusecase.ErrSlotUnavailable):
			bh.renderCalendar(w, r, input.Date, input, "Это время уже занято. Выберите другой слот.", http.StatusConflict)
		case errors.Is(err, bookingusecase.ErrInvalidBooking):
			bh.renderCalendar(w, r, input.Date, input, "Проверьте правильность заполнения формы.", http.StatusBadRequest)
		default:
			log.Println("Error creating booking:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	data := bookingdto.NewBookingSuccessDto(userId, host, created, input.Date, input.Time)
	data.Id = authenticatedUserID(r)
	t, err := template.ParseFiles("web/template/header.html", "web/template/booking_success.html", "web/template/footer.html")
	if err != nil {
		log.Println("Error parsing booking success template:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var page bytes.Buffer
	if err := t.ExecuteTemplate(&page, "booking_success", data); err != nil {
		log.Println("Error executing booking success template:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	if _, err := page.WriteTo(w); err != nil {
		log.Println("Error writing booking success page:", err)
	}
}

func (bh *BookingHandler) renderCalendar(
	w http.ResponseWriter,
	r *http.Request,
	date string,
	form bookingusecase.CreateInput,
	errorMessage string,
	status int,
) {
	userId := chi.URLParam(r, "id")
	availability, err := bh.BookingUC.Availability(r.Context(), userId, date)
	if err != nil {
		switch {
		case errors.Is(err, bookingusecase.ErrCalendarNotFound):
			http.Error(w, "Calendar not found", http.StatusNotFound)
		case errors.Is(err, bookingusecase.ErrInvalidBooking):
			http.Error(w, "Invalid date", http.StatusBadRequest)
		default:
			log.Println("Error getting calendar availability:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	host, err := bh.hostName(r, userId)
	if err != nil {
		if errors.Is(err, bookingusecase.ErrCalendarNotFound) {
			http.Error(w, "Calendar not found", http.StatusNotFound)
			return
		}
		log.Println("Error getting calendar owner:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := bookingdto.NewCalendarPageDto(
		userId,
		host,
		availability,
		time.Now(),
		form,
		errorMessage,
	)
	data.Id = authenticatedUserID(r)
	t, err := template.ParseFiles("web/template/header.html", "web/template/booking_calendar.html", "web/template/footer.html")
	if err != nil {
		log.Println("Error parsing booking calendar template:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var page bytes.Buffer
	if err := t.ExecuteTemplate(&page, "booking_calendar", data); err != nil {
		log.Println("Error executing booking calendar template:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if _, err := page.WriteTo(w); err != nil {
		log.Println("Error writing booking calendar page:", err)
	}
}

func (bh *BookingHandler) hostName(r *http.Request, userId string) (string, error) {
	host, err := bh.UserUC.UserRepository.GetById(r.Context(), userId)
	if err != nil {
		return "", err
	}
	if host == nil {
		return "", bookingusecase.ErrCalendarNotFound
	}
	return host.Name, nil
}

func authenticatedUserID(r *http.Request) string {
	userID, _ := r.Context().Value(middleware.UserKey).(string)
	return userID
}
