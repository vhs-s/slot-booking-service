package dto

import (
	"time"

	bookingentity "slot-booking-service-api/internal/entities/booking"
	bookingusecase "slot-booking-service-api/internal/usecase/booking"
)

type CalendarPageDto struct {
	Id             string
	CalendarUserId string
	HostName       string
	Title          string
	Timezone       string
	Date           string
	MinDate        string
	MaxDate        string
	Slots          []CalendarSlotDto
	Form           bookingusecase.CreateInput
	Error          string
}

type CalendarSlotDto struct {
	Value string
	Label string
}

func NewCalendarPageDto(
	userId string,
	hostName string,
	availability *bookingusecase.DayAvailability,
	minDate time.Time,
	form bookingusecase.CreateInput,
	errorMessage string,
) *CalendarPageDto {
	location := availability.Date.Location()
	page := &CalendarPageDto{
		CalendarUserId: userId,
		HostName:       hostName,
		Title:          availability.Calendar.Title,
		Timezone:       availability.Calendar.Timezone,
		Date:           availability.Date.Format("2006-01-02"),
		MinDate:        minDate.In(location).Format("2006-01-02"),
		MaxDate:        minDate.In(location).AddDate(0, 0, 90).Format("2006-01-02"),
		Slots:          make([]CalendarSlotDto, 0, len(availability.Slots)),
		Form:           form,
		Error:          errorMessage,
	}
	page.Form.Date = page.Date

	for _, slot := range availability.Slots {
		page.Slots = append(page.Slots, CalendarSlotDto{
			Value: slot.StartAt.In(location).Format("15:04"),
			Label: slot.StartAt.In(location).Format("15:04"),
		})
	}
	return page
}

type BookingSuccessDto struct {
	Id             string
	CalendarUserId string
	HostName       string
	Booking        *bookingentity.Booking
	DateTime       string
}

func NewBookingSuccessDto(userId, hostName string, booking *bookingentity.Booking, date, startTime string) *BookingSuccessDto {
	selectedTime, _ := time.Parse("2006-01-02 15:04", date+" "+startTime)
	return &BookingSuccessDto{
		CalendarUserId: userId,
		HostName:       hostName,
		Booking:        booking,
		DateTime:       selectedTime.Format("02.01.2006 15:04"),
	}
}
