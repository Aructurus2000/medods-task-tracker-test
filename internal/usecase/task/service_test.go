package task

import (
	"errors"
	"testing"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func TestValidateRecurrence(t *testing.T) {
	tests := []struct {
		name    string
		input   taskdomain.Recurrence
		wantErr bool
		check   func(t *testing.T, got taskdomain.Recurrence)
	}{
		{
			name:  "empty type defaults to none",
			input: taskdomain.Recurrence{},
			check: func(t *testing.T, got taskdomain.Recurrence) {
				t.Helper()
				if got.Type != taskdomain.RecurrenceNone {
					t.Fatalf("expected type=%s, got %s", taskdomain.RecurrenceNone, got.Type)
				}
			},
		},
		{
			name:  "explicit none is valid",
			input: taskdomain.Recurrence{Type: taskdomain.RecurrenceNone},
			check: func(t *testing.T, got taskdomain.Recurrence) {
				t.Helper()
				if got.Type != taskdomain.RecurrenceNone {
					t.Fatalf("expected type=%s, got %s", taskdomain.RecurrenceNone, got.Type)
				}
			},
		},
		{
			name:  "every_n_days one is valid",
			input: taskdomain.Recurrence{Type: taskdomain.RecurrenceEveryNDay, EveryNDays: 1},
		},
		{
			name:    "every_n_days zero is invalid",
			input:   taskdomain.Recurrence{Type: taskdomain.RecurrenceEveryNDay, EveryNDays: 0},
			wantErr: true,
		},
		{
			name:    "every_n_days negative is invalid",
			input:   taskdomain.Recurrence{Type: taskdomain.RecurrenceEveryNDay, EveryNDays: -1},
			wantErr: true,
		},
		{
			name:  "monthly day 1 is valid",
			input: taskdomain.Recurrence{Type: taskdomain.RecurrenceMonthly, DayOfMonth: 1},
		},
		{
			name:  "monthly day 31 is valid",
			input: taskdomain.Recurrence{Type: taskdomain.RecurrenceMonthly, DayOfMonth: 31},
		},
		{
			name:    "monthly day 0 is invalid",
			input:   taskdomain.Recurrence{Type: taskdomain.RecurrenceMonthly, DayOfMonth: 0},
			wantErr: true,
		},
		{
			name:    "monthly day 32 is invalid",
			input:   taskdomain.Recurrence{Type: taskdomain.RecurrenceMonthly, DayOfMonth: 32},
			wantErr: true,
		},
		{
			name: "specific_dates two dates valid",
			input: taskdomain.Recurrence{
				Type: taskdomain.RecurrenceDates,
				SpecificDates: []time.Time{
					time.Date(2026, time.April, 10, 15, 0, 0, 0, time.FixedZone("x", 3*3600)),
					time.Date(2026, time.April, 11, 0, 0, 0, 0, time.UTC),
				},
			},
			check: func(t *testing.T, got taskdomain.Recurrence) {
				t.Helper()
				if len(got.SpecificDates) != 2 {
					t.Fatalf("expected 2 dates, got %d", len(got.SpecificDates))
				}
			},
		},
		{
			name: "specific_dates deduplicates same day",
			input: taskdomain.Recurrence{
				Type: taskdomain.RecurrenceDates,
				SpecificDates: []time.Time{
					time.Date(2026, time.April, 10, 0, 0, 0, 0, time.UTC),
					time.Date(2026, time.April, 10, 15, 0, 0, 0, time.UTC),
				},
			},
			check: func(t *testing.T, got taskdomain.Recurrence) {
				t.Helper()
				if len(got.SpecificDates) != 1 {
					t.Fatalf("expected 1 deduplicated date, got %d", len(got.SpecificDates))
				}
			},
		},
		{
			name:    "specific_dates empty invalid",
			input:   taskdomain.Recurrence{Type: taskdomain.RecurrenceDates, SpecificDates: []time.Time{}},
			wantErr: true,
		},
		{
			name:  "parity even valid",
			input: taskdomain.Recurrence{Type: taskdomain.RecurrenceParity, Parity: taskdomain.DayParityEven},
		},
		{
			name:  "parity odd valid",
			input: taskdomain.Recurrence{Type: taskdomain.RecurrenceParity, Parity: taskdomain.DayParityOdd},
		},
		{
			name:    "parity wrong invalid",
			input:   taskdomain.Recurrence{Type: taskdomain.RecurrenceParity, Parity: taskdomain.DayParity("wrong")},
			wantErr: true,
		},
		{
			name:    "invalid type weekly",
			input:   taskdomain.Recurrence{Type: taskdomain.RecurrenceType("weekly")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateRecurrence(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidInput) {
					t.Fatalf("expected ErrInvalidInput, got %v", err)
				}
				return
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}
