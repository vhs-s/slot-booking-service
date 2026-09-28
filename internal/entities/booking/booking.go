package booking

import (
	"time"
)

type BookingStatus string

const (
	BookingStatusBooked    BookingStatus = "booked"
	BookingStatusCancelled BookingStatus = "cancelled"
)

type Booking struct {
	Id         string
	CalendarId string

	GuestName  string
	GuestEmail string

	Topic   string
	Comment string

	MeetingURL string

	StartAt time.Time
	EndAt   time.Time

	Status BookingStatus

	CreatedAt time.Time
}
