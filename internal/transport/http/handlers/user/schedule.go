package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"

	dto "slot-booking-service-api/internal/dto/schedule"
	"slot-booking-service-api/internal/transport/http/middleware"
	scheduleusecase "slot-booking-service-api/internal/usecase/schedule"

	"github.com/go-chi/chi/v5"
)

func (uh *UserHandler) ScheduleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, ok := r.Context().Value(middleware.UserKey).(string)
		if !ok || userId == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if pathUserId := chi.URLParam(r, "id"); pathUserId != "" && pathUserId != userId {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		switch r.Method {
		case http.MethodGet:
			uh.getSchedule(w, r, userId)
		case http.MethodPost:
			uh.saveSchedule(w, r, userId)
		default:
			w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (uh *UserHandler) getSchedule(w http.ResponseWriter, r *http.Request, userId string) {
	if uh.ScheduleUC == nil {
		http.Error(w, "Schedule service is not configured", http.StatusInternalServerError)
		return
	}

	userCalendar, rules, err := uh.ScheduleUC.Get(r.Context(), userId)
	if err != nil {
		log.Println("Error getting schedule:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := dto.NewGetScheduleDto(userId, userCalendar, rules)
	t, err := template.ParseFiles("web/template/header.html", "web/template/schedule.html", "web/template/footer.html")
	if err != nil {
		log.Println("Error parsing schedule template:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var page bytes.Buffer
	if err := t.ExecuteTemplate(&page, "schedule", data); err != nil {
		log.Println("Error executing schedule template:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := page.WriteTo(w); err != nil {
		log.Println("Error writing schedule page:", err)
	}
}

func (uh *UserHandler) saveSchedule(w http.ResponseWriter, r *http.Request, userId string) {
	if uh.ScheduleUC == nil {
		http.Error(w, "Schedule service is not configured", http.StatusInternalServerError)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	input, err := parseScheduleInput(r.PostForm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := uh.ScheduleUC.Save(r.Context(), userId, input); err != nil {
		if errors.Is(err, scheduleusecase.ErrInvalidSchedule) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Println("Error saving schedule:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/api/users/%s/schedule", userId), http.StatusSeeOther)
}

func parseScheduleInput(form url.Values) (scheduleusecase.SaveInput, error) {
	slotDuration, err := strconv.Atoi(form.Get("slot_duration"))
	if err != nil {
		return scheduleusecase.SaveInput{}, fmt.Errorf("invalid slot duration")
	}

	weekdays := form["weekday"]
	startTimes := form["start_time"]
	endTimes := form["end_time"]
	if len(weekdays) != len(startTimes) || len(weekdays) != len(endTimes) {
		return scheduleusecase.SaveInput{}, fmt.Errorf("invalid availability rules")
	}

	input := scheduleusecase.SaveInput{
		SlotDuration: slotDuration,
		Timezone:     form.Get("timezone"),
		Rules:        make([]scheduleusecase.RuleInput, 0, len(weekdays)),
	}
	for i := range weekdays {
		weekday, err := strconv.Atoi(weekdays[i])
		if err != nil {
			return scheduleusecase.SaveInput{}, fmt.Errorf("invalid weekday")
		}
		input.Rules = append(input.Rules, scheduleusecase.RuleInput{
			Weekday:   weekday,
			StartTime: startTimes[i],
			EndTime:   endTimes[i],
		})
	}

	return input, nil
}
