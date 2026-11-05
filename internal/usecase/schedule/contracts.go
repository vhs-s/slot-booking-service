package schedule

import (
	"context"
	availabilityrule "slot-booking-service-api/internal/entities/availabilityRule"
	"slot-booking-service-api/internal/entities/calendar"
)

type ScheduleRepo interface {
	GetByUserId(ctx context.Context, userId string) (*calendar.Calendar, []availabilityrule.AvailabilityRule, error)
	Save(ctx context.Context, calendar *calendar.Calendar, rules []availabilityrule.AvailabilityRule) error
}
