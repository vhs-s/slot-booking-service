package schedule

import (
	"context"
	"errors"
	"time"

	availabilityrule "slot-booking-service-api/internal/entities/availabilityRule"
	"slot-booking-service-api/internal/entities/calendar"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *repository {
	return &repository{db: db}
}

func (r *repository) GetByUserId(ctx context.Context, userId string) (*calendar.Calendar, []availabilityrule.AvailabilityRule, error) {
	var userCalendar calendar.Calendar
	var slotDuration int

	err := r.db.QueryRow(ctx, GET_CALENDAR_BY_USER_ID, userId).Scan(
		&userCalendar.Id,
		&userCalendar.UserId,
		&userCalendar.Title,
		&userCalendar.Timezone,
		&slotDuration,
		&userCalendar.IsActive,
		&userCalendar.CreatedAt,
		&userCalendar.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, []availabilityrule.AvailabilityRule{}, nil
	}
	if err != nil {
		return nil, nil, err
	}
	userCalendar.SlotDuration = time.Duration(slotDuration) * time.Minute

	rows, err := r.db.Query(ctx, GET_RULES_BY_CALENDAR_ID, userCalendar.Id)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	rules := make([]availabilityrule.AvailabilityRule, 0)
	for rows.Next() {
		var rule availabilityrule.AvailabilityRule
		if err := rows.Scan(&rule.Id, &rule.CalendarId, &rule.Weekday, &rule.StartTime, &rule.EndTime); err != nil {
			return nil, nil, err
		}
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return &userCalendar, rules, nil
}

func (r *repository) Save(ctx context.Context, userCalendar *calendar.Calendar, rules []availabilityrule.AvailabilityRule) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var calendarId string
	err = tx.QueryRow(ctx, UPSERT_CALENDAR,
		userCalendar.Id,
		userCalendar.UserId,
		userCalendar.Title,
		userCalendar.Timezone,
		int(userCalendar.SlotDuration.Minutes()),
		userCalendar.IsActive,
		userCalendar.CreatedAt,
		userCalendar.UpdatedAt,
	).Scan(&calendarId)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, DELETE_RULES_BY_CALENDAR_ID, calendarId); err != nil {
		return err
	}

	for _, rule := range rules {
		if _, err := tx.Exec(ctx, CREATE_RULE,
			rule.Id,
			calendarId,
			rule.Weekday,
			rule.StartTime,
			rule.EndTime,
		); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
