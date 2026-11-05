package dto

import (
	"fmt"
	"time"

	userdto "slot-booking-service-api/internal/dto/user"
	bookingentity "slot-booking-service-api/internal/entities/booking"
)

type MeetingStatusFilter string

const (
	MeetingFilterUpcoming  MeetingStatusFilter = "upcoming"
	MeetingFilterCompleted MeetingStatusFilter = "completed"
	MeetingFilterCancelled MeetingStatusFilter = "cancelled"
	MeetingFilterAll       MeetingStatusFilter = "all"
)

type DashboardDto struct {
	*userdto.GetUserDto
	Meetings      []MeetingDto
	StatusOptions []MeetingStatusOption
	EmptyMessage  string
}

type MeetingStatusOption struct {
	Value    MeetingStatusFilter
	Label    string
	Selected bool
}

type MeetingDto struct {
	GuestName  string
	GuestEmail string
	Topic      string
	Comment    string
	JoinURL    string
	Date       string
	StartTime  string
	EndTime    string
	Status     string
	CanJoin    bool
}

type meetingState string

const (
	meetingStateBooked    meetingState = "booked"
	meetingStateOngoing   meetingState = "ongoing"
	meetingStateCompleted meetingState = "completed"
	meetingStateCancelled meetingState = "cancelled"
)

func ParseMeetingStatusFilter(value string) MeetingStatusFilter {
	switch MeetingStatusFilter(value) {
	case MeetingFilterCompleted, MeetingFilterCancelled, MeetingFilterAll:
		return MeetingStatusFilter(value)
	default:
		return MeetingFilterUpcoming
	}
}

func NewDashboardDto(
	user *userdto.GetUserDto,
	bookings []bookingentity.Booking,
	location *time.Location,
	filter MeetingStatusFilter,
	now time.Time,
) *DashboardDto {
	filter = ParseMeetingStatusFilter(string(filter))
	dashboard := &DashboardDto{
		GetUserDto:   user,
		Meetings:     make([]MeetingDto, 0, len(bookings)),
		EmptyMessage: "По выбранному фильтру встреч нет.",
	}
	if filter == MeetingFilterUpcoming {
		dashboard.EmptyMessage = "У вас нет предстоящих встреч."
	}

	for _, option := range []MeetingStatusOption{
		{Value: MeetingFilterUpcoming, Label: "Предстоящие"},
		{Value: MeetingFilterCompleted, Label: "Завершённые"},
		{Value: MeetingFilterCancelled, Label: "Отменённые"},
		{Value: MeetingFilterAll, Label: "Все"},
	} {
		option.Selected = option.Value == filter
		dashboard.StatusOptions = append(dashboard.StatusOptions, option)
	}

	for _, booking := range bookings {
		state := getMeetingState(booking, now)
		if !matchesMeetingFilter(state, filter) {
			continue
		}

		status := "Запланирована"
		switch state {
		case meetingStateOngoing:
			status = "Идёт сейчас"
		case meetingStateCompleted:
			status = "Завершена"
		case meetingStateCancelled:
			status = "Отменена"
		}

		dashboard.Meetings = append(dashboard.Meetings, MeetingDto{
			GuestName:  booking.GuestName,
			GuestEmail: booking.GuestEmail,
			Topic:      booking.Topic,
			Comment:    booking.Comment,
			JoinURL:    fmt.Sprintf("/api/users/%s/meetings/%s/join", user.Id, booking.Id),
			Date:       booking.StartAt.In(location).Format("02.01.2006"),
			StartTime:  booking.StartAt.In(location).Format("15:04"),
			EndTime:    booking.EndAt.In(location).Format("15:04"),
			Status:     status,
			CanJoin:    state == meetingStateBooked || state == meetingStateOngoing,
		})
	}
	return dashboard
}

func getMeetingState(booking bookingentity.Booking, now time.Time) meetingState {
	if booking.Status == bookingentity.BookingStatusCancelled {
		return meetingStateCancelled
	}
	if !booking.EndAt.After(now) {
		return meetingStateCompleted
	}
	if !booking.StartAt.After(now) {
		return meetingStateOngoing
	}
	return meetingStateBooked
}

func matchesMeetingFilter(state meetingState, filter MeetingStatusFilter) bool {
	switch filter {
	case MeetingFilterCompleted:
		return state == meetingStateCompleted
	case MeetingFilterCancelled:
		return state == meetingStateCancelled
	case MeetingFilterAll:
		return true
	default:
		return state == meetingStateBooked || state == meetingStateOngoing
	}
}
