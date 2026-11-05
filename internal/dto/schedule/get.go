package dto

import (
	availabilityrule "slot-booking-service-api/internal/entities/availabilityRule"
	"slot-booking-service-api/internal/entities/calendar"
)

type GetScheduleDto struct {
	Id           string
	SlotDuration int
	Timezone     string
	Days         []ScheduleDayDto
}

type ScheduleDayDto struct {
	Weekday int
	Name    string
	Enabled bool
	Rules   []ScheduleRuleDto
}

type ScheduleRuleDto struct {
	StartTime string
	EndTime   string
}

var weekdayNames = []string{
	"Понедельник",
	"Вторник",
	"Среда",
	"Четверг",
	"Пятница",
	"Суббота",
	"Воскресенье",
}

func NewGetScheduleDto(userId string, userCalendar *calendar.Calendar, rules []availabilityrule.AvailabilityRule) *GetScheduleDto {
	schedule := &GetScheduleDto{
		Id:           userId,
		SlotDuration: 30,
		Timezone:     "UTC",
		Days:         make([]ScheduleDayDto, len(weekdayNames)),
	}

	if userCalendar != nil {
		schedule.SlotDuration = int(userCalendar.SlotDuration.Minutes())
		schedule.Timezone = userCalendar.Timezone
	}

	for i, name := range weekdayNames {
		schedule.Days[i] = ScheduleDayDto{
			Weekday: i + 1,
			Name:    name,
		}
	}

	for _, rule := range rules {
		if rule.Weekday < 1 || rule.Weekday > len(schedule.Days) {
			continue
		}

		day := &schedule.Days[rule.Weekday-1]
		day.Enabled = true
		day.Rules = append(day.Rules, ScheduleRuleDto{
			StartTime: rule.StartTime.Format("15:04"),
			EndTime:   rule.EndTime.Format("15:04"),
		})
	}

	return schedule
}
