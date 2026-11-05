CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY,
    calendar_id UUID NOT NULL REFERENCES calendars(id) ON DELETE CASCADE,
    guest_name TEXT NOT NULL,
    guest_email VARCHAR(254) NOT NULL,
    topic TEXT NOT NULL,
    comment TEXT NOT NULL DEFAULT '',
    meeting_url TEXT NOT NULL,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'booked',
    created_at TIMESTAMPTZ NOT NULL,
    CHECK (end_at > start_at),
    CHECK (status IN ('booked', 'cancelled'))
);

CREATE INDEX IF NOT EXISTS bookings_calendar_id_time_idx
    ON bookings(calendar_id, start_at, end_at);
