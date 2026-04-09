package jobs

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"example.com/taskservice/internal/usecase/reminder"
)

// RunMorningReminderLoop wakes up at REMINDER_AT_UTC (default 08:00) once per calendar day (UTC) and runs digest.
func RunMorningReminderLoop(ctx context.Context, logger *slog.Logger, svc *reminder.Service, reminderAtUTC string) {
	reminderAtUTC = strings.TrimSpace(reminderAtUTC)
	if reminderAtUTC == "" {
		reminderAtUTC = "08:00"
	}
	hour, min := parseHHMM(reminderAtUTC)
	if hour < 0 {
		logger.Error("invalid REMINDER_AT_UTC, use HH:MM", "value", reminderAtUTC)
		return
	}

	var mu sync.Mutex
	var lastRunDate string

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now().UTC()
			dateKey := now.Format("2006-01-02")
			if now.Hour() != hour || now.Minute() != min {
				continue
			}
			mu.Lock()
			if lastRunDate == dateKey {
				mu.Unlock()
				continue
			}
			lastRunDate = dateKey
			mu.Unlock()

			n, err := svc.RunMorningDigest(ctx, now)
			if err != nil {
				logger.Error("morning reminder digest failed", "error", err)
				continue
			}
			logger.Info("morning reminder digest done", "date", dateKey, "notifications_created", n)
		}
	}
}

func parseHHMM(s string) (hour, min int) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return -1, -1
	}
	h, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return -1, -1
	}
	m, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return -1, -1
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return -1, -1
	}
	return h, m
}
