package calendar

import (
	"time"
)

type Calendar struct {
	Id     string
	UserId string

	Title string

	Timezone     string
	SlotDuration time.Duration
	IsActive     bool

	CreatedAt time.Time
	UpdatedAt time.Time
}
