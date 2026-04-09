CREATE TABLE IF NOT EXISTS tasks (
	id BIGSERIAL PRIMARY KEY,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL,
	recurrence_type TEXT NOT NULL DEFAULT 'none',
	recurrence_every_n_days INT NULL,
	recurrence_day_of_month INT NULL,
	recurrence_specific_dates JSONB NOT NULL DEFAULT '[]'::jsonb,
	recurrence_parity TEXT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks (status);
CREATE INDEX IF NOT EXISTS idx_tasks_recurrence_type ON tasks (recurrence_type);
