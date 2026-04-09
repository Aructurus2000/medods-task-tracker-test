package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	notificationdomain "example.com/taskservice/internal/domain/notification"
)

type NotificationLister interface {
	ListNotificationsByDate(ctx context.Context, day time.Time) ([]notificationdomain.Notification, error)
}

type NotificationHandler struct {
	repo NotificationLister
}

func NewNotificationHandler(repo NotificationLister) *NotificationHandler {
	return &NotificationHandler{repo: repo}
}

type notificationDTO struct {
	ID           int64     `json:"id"`
	TaskID       int64     `json:"task_id"`
	ReminderDate time.Time `json:"reminder_date"`
	Message      string    `json:"message"`
	CreatedAt    time.Time `json:"created_at"`
}

func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	day, err := parseNotificationDateQuery(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	items, err := h.repo.ListNotificationsByDate(r.Context(), day)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	out := make([]notificationDTO, 0, len(items))
	for i := range items {
		out = append(out, notificationDTO{
			ID:           items[i].ID,
			TaskID:       items[i].TaskID,
			ReminderDate: items[i].ReminderDate,
			Message:      items[i].Message,
			CreatedAt:    items[i].CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func parseNotificationDateQuery(r *http.Request) (time.Time, error) {
	raw := r.URL.Query().Get("date")
	if raw == "" {
		now := time.Now().UTC()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC), nil
	}
	d, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, errors.New("invalid date query parameter, expected YYYY-MM-DD")
	}
	return time.Date(d.UTC().Year(), d.UTC().Month(), d.UTC().Day(), 0, 0, 0, 0, time.UTC), nil
}
