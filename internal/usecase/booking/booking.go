package booking

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	availabilityrule "slot-booking-service-api/internal/entities/availabilityRule"
	bookingentity "slot-booking-service-api/internal/entities/booking"
	"slot-booking-service-api/internal/entities/calendar"
	scheduleusecase "slot-booking-service-api/internal/usecase/schedule"

	"github.com/google/uuid"
)

var (
	ErrCalendarNotFound = errors.New("calendar not found")
	ErrInvalidBooking   = errors.New("invalid booking")
	ErrSlotUnavailable  = bookingentity.ErrSlotUnavailable
	ErrMeetingNotFound  = errors.New("meeting not found")
	ErrMeetingFinished  = errors.New("meeting is finished")
	ErrMeetingCancelled = errors.New("meeting is cancelled")
)

type TimeSlot struct {
	StartAt time.Time
	EndAt   time.Time
}

type DayAvailability struct {
	Calendar *calendar.Calendar
	Date     time.Time
	Slots    []TimeSlot
}

type CreateInput struct {
	Date       string
	Time       string
	GuestName  string
	GuestEmail string
	Topic      string
	Comment    string
	MeetingURL string
}

type BookingUseCase struct {
	BookingRepository BookingRepo
	ScheduleUC        *scheduleusecase.ScheduleUseCase
	now               func() time.Time
}

func New(bookingRepo BookingRepo, scheduleUC *scheduleusecase.ScheduleUseCase) *BookingUseCase {
	return &BookingUseCase{
		BookingRepository: bookingRepo,
		ScheduleUC:        scheduleUC,
		now:               time.Now,
	}
}

func (b *BookingUseCase) Availability(ctx context.Context, userId, date string) (*DayAvailability, error) {
	userCalendar, rules, err := b.ScheduleUC.Get(ctx, userId)
	if err != nil {
		return nil, err
	}
	if userCalendar == nil || !userCalendar.IsActive {
		return nil, ErrCalendarNotFound
	}

	location, err := time.LoadLocation(userCalendar.Timezone)
	if err != nil {
		return nil, fmt.Errorf("load calendar timezone: %w", err)
	}

	var selectedDate time.Time
	if strings.TrimSpace(date) == "" {
		now := b.now().In(location)
		selectedDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	} else {
		selectedDate, err = time.ParseInLocation("2006-01-02", strings.TrimSpace(date), location)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid date", ErrInvalidBooking)
		}
	}

	dayEnd := selectedDate.AddDate(0, 0, 1)
	existing, err := b.BookingRepository.ListForCalendarBetween(
		ctx,
		userCalendar.Id,
		selectedDate.UTC(),
		dayEnd.UTC(),
	)
	if err != nil {
		return nil, err
	}

	return &DayAvailability{
		Calendar: userCalendar,
		Date:     selectedDate,
		Slots:    availableSlots(userCalendar, rules, existing, selectedDate, b.now()),
	}, nil
}

func (b *BookingUseCase) Create(ctx context.Context, userId string, input CreateInput) (*bookingentity.Booking, error) {
	input.Date = strings.TrimSpace(input.Date)
	input.Time = strings.TrimSpace(input.Time)
	input.GuestName = strings.TrimSpace(input.GuestName)
	input.GuestEmail = strings.TrimSpace(input.GuestEmail)
	input.Topic = strings.TrimSpace(input.Topic)
	input.Comment = strings.TrimSpace(input.Comment)
	input.MeetingURL = strings.TrimSpace(input.MeetingURL)

	if err := validateCreateInput(input); err != nil {
		return nil, err
	}

	availability, err := b.Availability(ctx, userId, input.Date)
	if err != nil {
		return nil, err
	}

	location := availability.Date.Location()
	startAt, err := time.ParseInLocation("2006-01-02 15:04", input.Date+" "+input.Time, location)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid time", ErrInvalidBooking)
	}

	var selectedSlot *TimeSlot
	for i := range availability.Slots {
		if availability.Slots[i].StartAt.Equal(startAt) {
			selectedSlot = &availability.Slots[i]
			break
		}
	}
	if selectedSlot == nil {
		return nil, ErrSlotUnavailable
	}

	now := b.now().UTC()
	newBooking := &bookingentity.Booking{
		Id:         uuid.NewString(),
		CalendarId: availability.Calendar.Id,
		GuestName:  input.GuestName,
		GuestEmail: input.GuestEmail,
		Topic:      input.Topic,
		Comment:    input.Comment,
		MeetingURL: input.MeetingURL,
		StartAt:    selectedSlot.StartAt.UTC(),
		EndAt:      selectedSlot.EndAt.UTC(),
		Status:     bookingentity.BookingStatusBooked,
		CreatedAt:  now,
	}

	if err := b.BookingRepository.Create(ctx, newBooking); err != nil {
		return nil, err
	}
	return newBooking, nil
}

func (b *BookingUseCase) ListByUserId(ctx context.Context, userId string) ([]bookingentity.Booking, error) {
	return b.BookingRepository.ListByUserId(ctx, userId)
}

func (b *BookingUseCase) JoinMeeting(ctx context.Context, userId, bookingId string) (string, error) {
	booking, err := b.BookingRepository.GetByIdForUser(ctx, bookingId, userId)
	if err != nil {
		return "", err
	}
	if booking == nil {
		return "", ErrMeetingNotFound
	}
	if booking.Status == bookingentity.BookingStatusCancelled {
		return "", ErrMeetingCancelled
	}
	if !booking.EndAt.After(b.now()) {
		return "", ErrMeetingFinished
	}
	return booking.MeetingURL, nil
}

func validateCreateInput(input CreateInput) error {
	switch {
	case input.GuestName == "":
		return fmt.Errorf("%w: guest name is required", ErrInvalidBooking)
	case utf8.RuneCountInString(input.GuestName) > 100:
		return fmt.Errorf("%w: guest name is too long", ErrInvalidBooking)
	case input.GuestEmail == "" || utf8.RuneCountInString(input.GuestEmail) > 254:
		return fmt.Errorf("%w: invalid guest email", ErrInvalidBooking)
	case input.Topic == "":
		return fmt.Errorf("%w: topic is required", ErrInvalidBooking)
	case utf8.RuneCountInString(input.Topic) > 200:
		return fmt.Errorf("%w: topic is too long", ErrInvalidBooking)
	case utf8.RuneCountInString(input.Comment) > 2000:
		return fmt.Errorf("%w: comment is too long", ErrInvalidBooking)
	case input.MeetingURL == "" || utf8.RuneCountInString(input.MeetingURL) > 2048:
		return fmt.Errorf("%w: invalid meeting URL", ErrInvalidBooking)
	}

	meetingURL, err := url.ParseRequestURI(input.MeetingURL)
	if err != nil || meetingURL.Host == "" || (meetingURL.Scheme != "http" && meetingURL.Scheme != "https") {
		return fmt.Errorf("%w: invalid meeting URL", ErrInvalidBooking)
	}

	address, err := mail.ParseAddress(input.GuestEmail)
	if err != nil || address.Address != input.GuestEmail {
		return fmt.Errorf("%w: invalid guest email", ErrInvalidBooking)
	}
	return nil
}

func availableSlots(
	userCalendar *calendar.Calendar,
	rules []availabilityrule.AvailabilityRule,
	existing []bookingentity.Booking,
	date time.Time,
	now time.Time,
) []TimeSlot {
	slots := make([]TimeSlot, 0)
	if userCalendar.SlotDuration <= 0 {
		return slots
	}

	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	for _, rule := range rules {
		if rule.Weekday != weekday {
			continue
		}

		ruleStart := time.Date(
			date.Year(), date.Month(), date.Day(),
			rule.StartTime.Hour(), rule.StartTime.Minute(), rule.StartTime.Second(),
			0, date.Location(),
		)
		ruleEnd := time.Date(
			date.Year(), date.Month(), date.Day(),
			rule.EndTime.Hour(), rule.EndTime.Minute(), rule.EndTime.Second(),
			0, date.Location(),
		)

		for startAt := ruleStart; !startAt.Add(userCalendar.SlotDuration).After(ruleEnd); startAt = startAt.Add(userCalendar.SlotDuration) {
			if !startAt.After(now) {
				continue
			}

			slot := TimeSlot{StartAt: startAt, EndAt: startAt.Add(userCalendar.SlotDuration)}
			if overlapsExisting(slot, existing) {
				continue
			}
			slots = append(slots, slot)
		}
	}
	return slots
}

func overlapsExisting(slot TimeSlot, existing []bookingentity.Booking) bool {
	for _, item := range existing {
		if item.Status != bookingentity.BookingStatusBooked {
			continue
		}
		if item.StartAt.Before(slot.EndAt) && slot.StartAt.Before(item.EndAt) {
			return true
		}
	}
	return false
}
