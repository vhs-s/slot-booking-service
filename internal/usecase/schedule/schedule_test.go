package schedule

import (
	"context"
	"errors"
	"testing"
	"time"

	availabilityrule "slot-booking-service-api/internal/entities/availabilityRule"
	"slot-booking-service-api/internal/entities/calendar"
)

func TestScheduleUseCaseSave(t *testing.T) {
	repoErr := errors.New("repository unavailable")
	tests := []struct {
		name    string
		change  func(*SaveInput)
		repoErr error
		wantErr error
	}{
		{"valid_unsorted_adjacent_rules", nil, nil, nil},
		{"unsupported_duration", func(in *SaveInput) { in.SlotDuration = 20 }, nil, ErrInvalidSchedule},
		{"invalid_timezone", func(in *SaveInput) { in.Timezone = "Invalid/Timezone" }, nil, ErrInvalidSchedule},
		{"invalid_weekday", func(in *SaveInput) { in.Rules[0].Weekday = 8 }, nil, ErrInvalidSchedule},
		{"invalid_time_format", func(in *SaveInput) { in.Rules[0].StartTime = "invalid" }, nil, ErrInvalidSchedule},
		{"end_before_start", func(in *SaveInput) { in.Rules[0].EndTime = "09:00" }, nil, ErrInvalidSchedule},
		{"overlapping_rules", func(in *SaveInput) { in.Rules[0].StartTime = "09:15" }, nil, ErrInvalidSchedule},
		{"repository_error", nil, repoErr, repoErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := SaveInput{
				SlotDuration: 30, Timezone: "UTC",
				Rules: []RuleInput{
					{Weekday: 1, StartTime: "09:30", EndTime: "10:00"},
					{Weekday: 1, StartTime: "09:00", EndTime: "09:30"},
				},
			}
			if tt.change != nil {
				tt.change(&input)
			}
			repo := &scheduleRepoStub{err: tt.repoErr}
			err := New(repo).Save(context.Background(), "user-id", input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if errors.Is(tt.wantErr, ErrInvalidSchedule) {
				if repo.calendar != nil {
					t.Error("invalid schedule was sent to repository")
				}
				return
			}
			cal := repo.calendar
			if cal == nil {
				t.Fatal("expected schedule to be passed to repository")
			}
			if cal.Id == "" || cal.UserId != "user-id" || !cal.IsActive || cal.Timezone != "UTC" || cal.SlotDuration != 30*time.Minute {
				t.Errorf("unexpected calendar: %+v", cal)
			}
			if len(repo.rules) != 2 {
				t.Fatalf("rules count = %d, want 2", len(repo.rules))
			}
			for i, wantStart := range []string{"09:00", "09:30"} {
				rule := repo.rules[i]
				if rule.Id == "" || rule.Weekday != 1 || rule.StartTime.Format("15:04") != wantStart || rule.EndTime.Sub(rule.StartTime) != 30*time.Minute {
					t.Errorf("unexpected rule at index %d: %+v", i, rule)
				}
			}
		})
	}
}

type scheduleRepoStub struct {
	ScheduleRepo
	calendar *calendar.Calendar
	rules    []availabilityrule.AvailabilityRule
	err      error
}

func (r *scheduleRepoStub) Save(_ context.Context, cal *calendar.Calendar, rules []availabilityrule.AvailabilityRule) error {
	r.calendar, r.rules = cal, rules
	return r.err
}
