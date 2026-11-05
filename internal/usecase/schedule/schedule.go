package schedule

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	availabilityrule "slot-booking-service-api/internal/entities/availabilityRule"
	"slot-booking-service-api/internal/entities/calendar"

	"github.com/google/uuid"
)

var ErrInvalidSchedule = errors.New("invalid schedule")

type RuleInput struct {
	Weekday   int
	StartTime string
	EndTime   string
}

type SaveInput struct {
	SlotDuration int
	Timezone     string
	Rules        []RuleInput
}

type ScheduleUseCase struct {
	ScheduleRepository ScheduleRepo
}

func New(scheduleRepo ScheduleRepo) *ScheduleUseCase {
	return &ScheduleUseCase{ScheduleRepository: scheduleRepo}
}

func (s *ScheduleUseCase) Get(ctx context.Context, userId string) (*calendar.Calendar, []availabilityrule.AvailabilityRule, error) {
	return s.ScheduleRepository.GetByUserId(ctx, userId)
}

func (s *ScheduleUseCase) Save(ctx context.Context, userId string, input SaveInput) error {
	switch input.SlotDuration {
	case 15, 30, 45, 60:
	default:
		return fmt.Errorf("%w: unsupported slot duration", ErrInvalidSchedule)
	}

	if _, err := time.LoadLocation(input.Timezone); err != nil {
		return fmt.Errorf("%w: invalid timezone", ErrInvalidSchedule)
	}

	sort.Slice(input.Rules, func(i, j int) bool {
		if input.Rules[i].Weekday == input.Rules[j].Weekday {
			return input.Rules[i].StartTime < input.Rules[j].StartTime
		}
		return input.Rules[i].Weekday < input.Rules[j].Weekday
	})

	rules := make([]availabilityrule.AvailabilityRule, 0, len(input.Rules))
	for i, ruleInput := range input.Rules {
		if ruleInput.Weekday < 1 || ruleInput.Weekday > 7 {
			return fmt.Errorf("%w: weekday must be between 1 and 7", ErrInvalidSchedule)
		}

		startTime, err := time.Parse("15:04", ruleInput.StartTime)
		if err != nil {
			return fmt.Errorf("%w: invalid start time", ErrInvalidSchedule)
		}
		endTime, err := time.Parse("15:04", ruleInput.EndTime)
		if err != nil {
			return fmt.Errorf("%w: invalid end time", ErrInvalidSchedule)
		}
		if !startTime.Before(endTime) {
			return fmt.Errorf("%w: end time must be after start time", ErrInvalidSchedule)
		}

		if i > 0 {
			previous := input.Rules[i-1]
			if previous.Weekday == ruleInput.Weekday && ruleInput.StartTime < previous.EndTime {
				return fmt.Errorf("%w: availability rules must not overlap", ErrInvalidSchedule)
			}
		}

		rules = append(rules, availabilityrule.AvailabilityRule{
			Id:        uuid.NewString(),
			Weekday:   ruleInput.Weekday,
			StartTime: startTime,
			EndTime:   endTime,
		})
	}

	now := time.Now().UTC()
	userCalendar := &calendar.Calendar{
		Id:           uuid.NewString(),
		UserId:       userId,
		Title:        "Мой календарь",
		Timezone:     input.Timezone,
		SlotDuration: time.Duration(input.SlotDuration) * time.Minute,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return s.ScheduleRepository.Save(ctx, userCalendar, rules)
}
