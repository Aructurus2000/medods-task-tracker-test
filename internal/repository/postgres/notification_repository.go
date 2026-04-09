package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	notificationdomain "example.com/taskservice/internal/domain/notification"
)

func (r *Repository) InsertReminderIfNotExists(ctx context.Context, taskID int64, reminderDate time.Time, message string) (bool, error) {
	const query = `
		INSERT INTO notifications (task_id, reminder_date, message)
		VALUES ($1, $2::date, $3)
		ON CONFLICT (task_id, reminder_date) DO NOTHING
		RETURNING id
	`
	row := r.pool.QueryRow(ctx, query, taskID, reminderDate.Format("2006-01-02"), message)
	var id int64
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *Repository) ListNotificationsByDate(ctx context.Context, day time.Time) ([]notificationdomain.Notification, error) {
	const query = `
		SELECT id, task_id, reminder_date, message, created_at
		FROM notifications
		WHERE reminder_date = $1::date
		ORDER BY id DESC
	`
	rows, err := r.pool.Query(ctx, query, day.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]notificationdomain.Notification, 0)
	for rows.Next() {
		var n notificationdomain.Notification
		var reminderDate time.Time
		if err := rows.Scan(&n.ID, &n.TaskID, &reminderDate, &n.Message, &n.CreatedAt); err != nil {
			return nil, err
		}
		n.ReminderDate = reminderDate.UTC()
		out = append(out, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
