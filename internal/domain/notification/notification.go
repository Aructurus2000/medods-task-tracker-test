package notification

import "time"

type Notification struct {
	ID           int64     `json:"id"`
	TaskID       int64     `json:"task_id"`
	ReminderDate time.Time `json:"reminder_date"`
	Message      string    `json:"message"`
	CreatedAt    time.Time `json:"created_at"`
}
