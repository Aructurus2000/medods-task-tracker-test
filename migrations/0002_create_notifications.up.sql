CREATE TABLE IF NOT EXISTS notifications (
	id BIGSERIAL PRIMARY KEY,
	task_id BIGINT NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
	reminder_date DATE NOT NULL,
	message TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE (task_id, reminder_date)
);

CREATE INDEX IF NOT EXISTS idx_notifications_reminder_date ON notifications (reminder_date);
