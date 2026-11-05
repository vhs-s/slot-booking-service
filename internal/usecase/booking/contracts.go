package booking

import (
	"context"
	"time"

	bookingentity "slot-booking-service-api/internal/entities/booking"
)

type BookingRepo interface {
	ListForCalendarBetween(ctx context.Context, calendarId string, from, to time.Time) ([]bookingentity.Booking, error)
	ListByUserId(ctx context.Context, userId string) ([]bookingentity.Booking, error)
	GetByIdForUser(ctx context.Context, bookingId, userId string) (*bookingentity.Booking, error)
	Create(ctx context.Context, booking *bookingentity.Booking) error
}
