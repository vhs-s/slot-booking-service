package availabilityrule

import "time"

type AvailabilityRule struct {
	Id         string
	CalendarId string

	Weekday int

	StartTime time.Time
	EndTime   time.Time
}
