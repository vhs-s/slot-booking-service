package schedule

const GET_CALENDAR_BY_USER_ID = `
	SELECT id, user_id, title, timezone, slot_duration, is_active, created_at, updated_at
	FROM calendars
	WHERE user_id = $1;`

const GET_RULES_BY_CALENDAR_ID = `
	SELECT id, calendar_id, weekday, start_time, end_time
	FROM availability_rules
	WHERE calendar_id = $1
	ORDER BY weekday, start_time;`

const UPSERT_CALENDAR = `
	INSERT INTO calendars (
		id, user_id, title, timezone, slot_duration, is_active, created_at, updated_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (user_id) DO UPDATE SET
		timezone = EXCLUDED.timezone,
		slot_duration = EXCLUDED.slot_duration,
		is_active = EXCLUDED.is_active,
		updated_at = EXCLUDED.updated_at
	RETURNING id;`

const DELETE_RULES_BY_CALENDAR_ID = `DELETE FROM availability_rules WHERE calendar_id = $1;`

const CREATE_RULE = `
	INSERT INTO availability_rules (id, calendar_id, weekday, start_time, end_time)
	VALUES ($1, $2, $3, $4, $5);`
