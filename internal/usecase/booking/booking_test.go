package booking

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	availabilityrule "slot-booking-service-api/internal/entities/availabilityRule"
	bookingentity "slot-booking-service-api/internal/entities/booking"
	"slot-booking-service-api/internal/entities/calendar"
	scheduleusecase "slot-booking-service-api/internal/usecase/schedule"
)

// Все проверки используют один фиксированный понедельник.
func testTime(hour, minute int) time.Time {
	return time.Date(2030, time.January, 7, hour, minute, 0, 0, time.UTC)
}

func validCreateInput() CreateInput {
	return CreateInput{
		Date: "2030-01-07", Time: "09:00",
		GuestName: "Иван", GuestEmail: "ivan@example.com",
		Topic: "Консультация", MeetingURL: "https://example.com/meeting",
	}
}

func TestAvailableSlots(t *testing.T) {
	first := TimeSlot{StartAt: testTime(9, 0), EndAt: testTime(9, 30)}
	second := TimeSlot{StartAt: testTime(9, 30), EndAt: testTime(10, 0)}
	tests := []struct {
		name     string
		duration time.Duration
		weekday  int
		now      time.Time
		existing []bookingentity.Booking
		want     []TimeSlot
	}{
		{"split_interval", 30 * time.Minute, 1, testTime(8, 0), nil, []TimeSlot{first, second}},
		{"different_weekday", 30 * time.Minute, 2, testTime(8, 0), nil, nil},
		{"exclude_started_slot", 30 * time.Minute, 1, testTime(9, 0), nil, []TimeSlot{second}},
		{"exclude_booked_slot", 30 * time.Minute, 1, testTime(8, 0), []bookingentity.Booking{
			{StartAt: first.StartAt, EndAt: first.EndAt, Status: bookingentity.BookingStatusBooked},
		}, []TimeSlot{second}},
		{"zero_duration", 0, 1, testTime(8, 0), nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := []availabilityrule.AvailabilityRule{
				{Weekday: tt.weekday, StartTime: testTime(9, 0), EndTime: testTime(10, 15)},
			}
			got := availableSlots(&calendar.Calendar{SlotDuration: tt.duration}, rules, tt.existing, testTime(0, 0), tt.now)
			if !slices.Equal(got, tt.want) {
				t.Errorf("slots = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOverlapsExisting(t *testing.T) {
	slot := TimeSlot{StartAt: testTime(9, 0), EndAt: testTime(9, 30)}
	tests := []struct {
		name     string
		existing []bookingentity.Booking
		want     bool
	}{
		{"no_bookings", nil, false},
		{"same_interval", []bookingentity.Booking{
			{StartAt: testTime(9, 0), EndAt: testTime(9, 30), Status: bookingentity.BookingStatusBooked},
		}, true},
		{"partial_overlap", []bookingentity.Booking{
			{StartAt: testTime(9, 15), EndAt: testTime(9, 45), Status: bookingentity.BookingStatusBooked},
		}, true},
		{"adjacent_intervals", []bookingentity.Booking{
			{StartAt: testTime(8, 30), EndAt: testTime(9, 0), Status: bookingentity.BookingStatusBooked},
			{StartAt: testTime(9, 30), EndAt: testTime(10, 0), Status: bookingentity.BookingStatusBooked},
		}, false},
		{"cancelled_booking", []bookingentity.Booking{
			{StartAt: testTime(9, 0), EndAt: testTime(9, 30), Status: bookingentity.BookingStatusCancelled},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := overlapsExisting(slot, tt.existing); got != tt.want {
				t.Errorf("overlapsExisting() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateCreateInput(t *testing.T) {
	tests := []struct {
		name    string
		change  func(*CreateInput)
		wantErr error
	}{
		{"valid", nil, nil},
		{"missing_name", func(in *CreateInput) { in.GuestName = "" }, ErrInvalidBooking},
		{"invalid_email", func(in *CreateInput) { in.GuestEmail = "invalid" }, ErrInvalidBooking},
		{"missing_topic", func(in *CreateInput) { in.Topic = "" }, ErrInvalidBooking},
		{"unsupported_url_scheme", func(in *CreateInput) { in.MeetingURL = "ftp://example.com/meeting" }, ErrInvalidBooking},
		{"name_at_limit", func(in *CreateInput) { in.GuestName = strings.Repeat("Я", 100) }, nil},
		{"name_over_limit", func(in *CreateInput) { in.GuestName = strings.Repeat("Я", 101) }, ErrInvalidBooking},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := validCreateInput()
			if tt.change != nil {
				tt.change(&input)
			}
			if err := validateCreateInput(input); !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestBookingUseCaseCreate(t *testing.T) {
	repoErr := errors.New("repository unavailable")
	tests := []struct {
		name    string
		change  func(*CreateInput, *calendar.Calendar, *bookingRepoStub)
		wantErr error
	}{
		{"valid_with_spaces", func(in *CreateInput, _ *calendar.Calendar, _ *bookingRepoStub) {
			in.GuestName = " Иван "
		}, nil},
		{"invalid_input", func(in *CreateInput, _ *calendar.Calendar, _ *bookingRepoStub) {
			in.GuestName = ""
		}, ErrInvalidBooking},
		{"inactive_calendar", func(_ *CreateInput, cal *calendar.Calendar, _ *bookingRepoStub) {
			cal.IsActive = false
		}, ErrCalendarNotFound},
		{"slot_already_booked", func(_ *CreateInput, _ *calendar.Calendar, repo *bookingRepoStub) {
			repo.existing = []bookingentity.Booking{
				{StartAt: testTime(9, 0), EndAt: testTime(9, 30), Status: bookingentity.BookingStatusBooked},
			}
		}, ErrSlotUnavailable},
		{"repository_error", func(_ *CreateInput, _ *calendar.Calendar, repo *bookingRepoStub) {
			repo.createErr = repoErr
		}, repoErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := validCreateInput()
			cal := &calendar.Calendar{Id: "calendar-id", Timezone: "UTC", SlotDuration: 30 * time.Minute, IsActive: true}
			repo := &bookingRepoStub{}
			tt.change(&input, cal, repo)
			scheduleRepo := &scheduleRepoStub{calendar: cal, rules: []availabilityrule.AvailabilityRule{
				{Weekday: 1, StartTime: testTime(9, 0), EndTime: testTime(10, 0)},
			}}
			uc := New(repo, scheduleusecase.New(scheduleRepo))
			uc.now = func() time.Time { return testTime(8, 0) }

			got, err := uc.Create(context.Background(), "user-id", input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if got != nil {
					t.Errorf("booking = %v, want nil on error", got)
				}
				if !errors.Is(tt.wantErr, repoErr) && repo.created != nil {
					t.Error("invalid booking was sent to repository")
				}
				return
			}
			if got == nil || got != repo.created {
				t.Fatal("expected created booking to be passed to repository and returned")
			}
			if got.Id == "" || got.CalendarId != cal.Id || got.Status != bookingentity.BookingStatusBooked {
				t.Errorf("unexpected booking metadata: %+v", got)
			}
			if got.GuestName != "Иван" || got.GuestEmail != input.GuestEmail || got.Topic != input.Topic || got.MeetingURL != input.MeetingURL {
				t.Errorf("unexpected booking details: %+v", got)
			}
			if !got.StartAt.Equal(testTime(9, 0)) || !got.EndAt.Equal(testTime(9, 30)) || !got.CreatedAt.Equal(testTime(8, 0)) {
				t.Errorf("unexpected booking times: %+v", got)
			}
		})
	}
}

func TestJoinMeeting(t *testing.T) {
	repoErr := errors.New("repository unavailable")
	tests := []struct {
		name    string
		booking *bookingentity.Booking
		repoErr error
		wantErr error
	}{
		{"active", &bookingentity.Booking{Status: bookingentity.BookingStatusBooked, EndAt: testTime(10, 0)}, nil, nil},
		{"not_found", nil, nil, ErrMeetingNotFound},
		{"cancelled", &bookingentity.Booking{Status: bookingentity.BookingStatusCancelled, EndAt: testTime(10, 0)}, nil, ErrMeetingCancelled},
		{"finished_at_current_time", &bookingentity.Booking{Status: bookingentity.BookingStatusBooked, EndAt: testTime(9, 0)}, nil, ErrMeetingFinished},
		{"repository_error", nil, repoErr, repoErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const meetingURL = "https://example.com/meeting"
			if tt.booking != nil {
				tt.booking.MeetingURL = meetingURL
			}
			repo := &bookingRepoStub{booking: tt.booking, getErr: tt.repoErr}
			uc := New(repo, nil)
			uc.now = func() time.Time { return testTime(9, 0) }
			got, err := uc.JoinMeeting(context.Background(), "user-id", "booking-id")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			wantURL := ""
			if tt.wantErr == nil {
				wantURL = meetingURL
			}
			if got != wantURL {
				t.Errorf("meeting URL = %q, want %q", got, wantURL)
			}
			if repo.userID != "user-id" || repo.bookingID != "booking-id" {
				t.Error("repository received incorrect user or booking ID")
			}
		})
	}
}

type bookingRepoStub struct {
	BookingRepo
	existing          []bookingentity.Booking
	booking, created  *bookingentity.Booking
	createErr, getErr error
	userID, bookingID string
}

func (r *bookingRepoStub) ListForCalendarBetween(context.Context, string, time.Time, time.Time) ([]bookingentity.Booking, error) {
	return r.existing, nil
}

func (r *bookingRepoStub) Create(_ context.Context, booking *bookingentity.Booking) error {
	r.created = booking
	return r.createErr
}

func (r *bookingRepoStub) GetByIdForUser(_ context.Context, bookingID, userID string) (*bookingentity.Booking, error) {
	r.bookingID, r.userID = bookingID, userID
	return r.booking, r.getErr
}

type scheduleRepoStub struct {
	scheduleusecase.ScheduleRepo
	calendar *calendar.Calendar
	rules    []availabilityrule.AvailabilityRule
}

func (r *scheduleRepoStub) GetByUserId(context.Context, string) (*calendar.Calendar, []availabilityrule.AvailabilityRule, error) {
	return r.calendar, r.rules, nil
}
