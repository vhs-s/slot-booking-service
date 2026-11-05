package booking

const LIST_FOR_CALENDAR_BETWEEN = `
	SELECT id, calendar_id, guest_name, guest_email, topic, comment,
		meeting_url, start_at, end_at, status, created_at
	FROM bookings
	WHERE calendar_id = $1
		AND status = 'booked'
		AND start_at < $3
		AND end_at > $2
	ORDER BY start_at;`

const LIST_BY_USER_ID = `
	SELECT b.id, b.calendar_id, b.guest_name, b.guest_email, b.topic, b.comment,
		b.meeting_url, b.start_at, b.end_at, b.status, b.created_at
	FROM bookings b
	JOIN calendars c ON c.id = b.calendar_id
	WHERE c.user_id = $1
	ORDER BY b.start_at;`

const GET_BY_ID_FOR_USER = `
	SELECT b.id, b.calendar_id, b.guest_name, b.guest_email, b.topic, b.comment,
		b.meeting_url, b.start_at, b.end_at, b.status, b.created_at
	FROM bookings b
	JOIN calendars c ON c.id = b.calendar_id
	WHERE b.id = $1
		AND c.user_id = $2;`

const LOCK_CALENDAR = `SELECT id FROM calendars WHERE id = $1 FOR UPDATE;`

const HAS_OVERLAPPING_BOOKING = `
	SELECT EXISTS (
		SELECT 1
		FROM bookings
		WHERE calendar_id = $1
			AND status = 'booked'
			AND start_at < $3
			AND end_at > $2
	);`

const CREATE_BOOKING = `
	INSERT INTO bookings (
		id, calendar_id, guest_name, guest_email, topic, comment,
		meeting_url, start_at, end_at, status, created_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`
