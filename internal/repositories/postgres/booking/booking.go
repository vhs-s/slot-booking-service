package booking

import (
	"context"
	"errors"
	"time"

	bookingentity "slot-booking-service-api/internal/entities/booking"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *repository {
	return &repository{db: db}
}

func (r *repository) ListForCalendarBetween(ctx context.Context, calendarId string, from, to time.Time) ([]bookingentity.Booking, error) {
	rows, err := r.db.Query(ctx, LIST_FOR_CALENDAR_BETWEEN, calendarId, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBookings(rows)
}

func (r *repository) ListByUserId(ctx context.Context, userId string) ([]bookingentity.Booking, error) {
	rows, err := r.db.Query(ctx, LIST_BY_USER_ID, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBookings(rows)
}

func (r *repository) GetByIdForUser(ctx context.Context, bookingId, userId string) (*bookingentity.Booking, error) {
	var item bookingentity.Booking
	err := r.db.QueryRow(ctx, GET_BY_ID_FOR_USER, bookingId, userId).Scan(
		&item.Id,
		&item.CalendarId,
		&item.GuestName,
		&item.GuestEmail,
		&item.Topic,
		&item.Comment,
		&item.MeetingURL,
		&item.StartAt,
		&item.EndAt,
		&item.Status,
		&item.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *repository) Create(ctx context.Context, booking *bookingentity.Booking) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Locking the calendar serializes booking creation for its owner. The
	// overlap check below therefore also protects against concurrent requests.
	var calendarId string
	if err := tx.QueryRow(ctx, LOCK_CALENDAR, booking.CalendarId).Scan(&calendarId); err != nil {
		return err
	}

	var overlaps bool
	if err := tx.QueryRow(ctx, HAS_OVERLAPPING_BOOKING,
		booking.CalendarId,
		booking.StartAt,
		booking.EndAt,
	).Scan(&overlaps); err != nil {
		return err
	}
	if overlaps {
		return bookingentity.ErrSlotUnavailable
	}

	if _, err := tx.Exec(ctx, CREATE_BOOKING,
		booking.Id,
		booking.CalendarId,
		booking.GuestName,
		booking.GuestEmail,
		booking.Topic,
		booking.Comment,
		booking.MeetingURL,
		booking.StartAt,
		booking.EndAt,
		booking.Status,
		booking.CreatedAt,
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

type bookingRows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

func scanBookings(rows bookingRows) ([]bookingentity.Booking, error) {
	bookings := make([]bookingentity.Booking, 0)
	for rows.Next() {
		var item bookingentity.Booking
		if err := rows.Scan(
			&item.Id,
			&item.CalendarId,
			&item.GuestName,
			&item.GuestEmail,
			&item.Topic,
			&item.Comment,
			&item.MeetingURL,
			&item.StartAt,
			&item.EndAt,
			&item.Status,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		bookings = append(bookings, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return bookings, nil
}
