package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	Recurrence  recurrenceDTO     `json:"recurrence"`
}

type recurrenceDTO struct {
	Type          taskdomain.RecurrenceType `json:"type"`
	EveryNDays    int                       `json:"every_n_days,omitempty"`
	DayOfMonth    int                       `json:"day_of_month,omitempty"`
	SpecificDates []string                  `json:"specific_dates,omitempty"`
	Parity        taskdomain.DayParity      `json:"parity,omitempty"`
}

type taskDTO struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	Recurrence  recurrenceDTO     `json:"recurrence"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Recurrence: recurrenceDTO{
			Type:          task.Recurrence.Type,
			EveryNDays:    task.Recurrence.EveryNDays,
			DayOfMonth:    task.Recurrence.DayOfMonth,
			SpecificDates: formatSpecificDates(task.Recurrence.SpecificDates),
			Parity:        task.Recurrence.Parity,
		},
		CreatedAt: task.CreatedAt,
		UpdatedAt: task.UpdatedAt,
	}
}

func formatSpecificDates(dates []time.Time) []string {
	if len(dates) == 0 {
		return nil
	}
	out := make([]string, 0, len(dates))
	for _, d := range dates {
		out = append(out, d.UTC().Format("2006-01-02"))
	}
	return out
}
