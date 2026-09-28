package handlers

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
)

type SchedulePageData struct { //test
	Name string
	Days []ScheduleDay
}

type ScheduleDay struct { //test
	Weekday int
	Name    string
	Enabled bool
	Rules   []ScheduleRule
}

type ScheduleRule struct { //test
	StartTime string
	EndTime   string
}

func (uh *UserHandler) ScheduleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			t, err := template.ParseFiles("web/template/header.html", "web/template/schedule.html", "web/template/footer.html")
			if err != nil {
				log.Println("Error parse template", err)
			}
			fmt.Println(r.Context().Value("userID"))

			data := SchedulePageData{ //test
				Name: "Vlad",

				Days: []ScheduleDay{
					{
						Weekday: 1,
						Name:    "Понедельник",
						Enabled: true,
						Rules: []ScheduleRule{
							{
								StartTime: "09:00",
								EndTime:   "12:00",
							},
							{
								StartTime: "13:00",
								EndTime:   "18:00",
							},
						},
					},
					{
						Weekday: 2,
						Name:    "Вторник",
						Enabled: true,
						Rules: []ScheduleRule{
							{
								StartTime: "09:00",
								EndTime:   "18:00",
							},
						},
					},
					{
						Weekday: 3,
						Name:    "Среда",
						Enabled: true,
						Rules: []ScheduleRule{
							{
								StartTime: "10:00",
								EndTime:   "13:00",
							},
							{
								StartTime: "14:00",
								EndTime:   "19:00",
							},
						},
					},
					{
						Weekday: 4,
						Name:    "Четверг",
						Enabled: true,
						Rules: []ScheduleRule{
							{
								StartTime: "09:00",
								EndTime:   "17:00",
							},
						},
					},
					{
						Weekday: 5,
						Name:    "Пятница",
						Enabled: true,
						Rules: []ScheduleRule{
							{
								StartTime: "09:00",
								EndTime:   "15:00",
							},
						},
					},
					{
						Weekday: 6,
						Name:    "Суббота",
						Enabled: false,
						Rules:   nil,
					},
					{
						Weekday: 7,
						Name:    "Воскресенье",
						Enabled: false,
						Rules:   nil,
					},
				},
			}
			t.ExecuteTemplate(w, "schedule", data)
			break
		case http.MethodPost:
		}
	}
}
