package reminder

import (
	"context"
	"fmt"
	"time"

	taskrepo "example.com/taskservice/internal/usecase/task"
)

type NotificationWriter interface {
	InsertReminderIfNotExists(ctx context.Context, taskID int64, reminderDate time.Time, message string) (bool, error)
}

type Service struct {
	tasks         taskrepo.Repository
	notifications NotificationWriter
}

func NewService(tasks taskrepo.Repository, notifications NotificationWriter) *Service {
	return &Service{tasks: tasks, notifications: notifications}
}

// RunMorningDigest creates one notification per open task that is due on reminderDate (UTC calendar day).
func (s *Service) RunMorningDigest(ctx context.Context, reminderDate time.Time) (created int, err error) {
	day := time.Date(reminderDate.UTC().Year(), reminderDate.UTC().Month(), reminderDate.UTC().Day(), 0, 0, 0, 0, time.UTC)
	tasks, err := s.tasks.ListOpenDueOnDate(ctx, day)
	if err != nil {
		return 0, err
	}
	for i := range tasks {
		t := &tasks[i]
		msg := fmt.Sprintf("Напоминание на %s: %s", day.Format("2006-01-02"), t.Title)
		inserted, err := s.notifications.InsertReminderIfNotExists(ctx, t.ID, day, msg)
		if err != nil {
			return created, err
		}
		if inserted {
			created++
		}
	}
	return created, nil
}
