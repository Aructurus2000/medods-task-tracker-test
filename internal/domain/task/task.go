package task

import "time"

type Status string
type RecurrenceType string
type DayParity string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

const (
	RecurrenceNone      RecurrenceType = "none"
	RecurrenceEveryNDay RecurrenceType = "every_n_days"
	RecurrenceMonthly   RecurrenceType = "monthly"
	RecurrenceDates     RecurrenceType = "specific_dates"
	RecurrenceParity    RecurrenceType = "month_day_parity"
)

const (
	DayParityEven DayParity = "even"
	DayParityOdd  DayParity = "odd"
)

type Recurrence struct {
	Type          RecurrenceType `json:"type"`
	EveryNDays    int            `json:"every_n_days,omitempty"`
	DayOfMonth    int            `json:"day_of_month,omitempty"`
	SpecificDates []time.Time    `json:"specific_dates,omitempty"`
	Parity        DayParity      `json:"parity,omitempty"`
}

type Task struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      Status     `json:"status"`
	Recurrence  Recurrence `json:"recurrence"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

func (r RecurrenceType) Valid() bool {
	switch r {
	case RecurrenceNone, RecurrenceEveryNDay, RecurrenceMonthly, RecurrenceDates, RecurrenceParity:
		return true
	default:
		return false
	}
}

func (p DayParity) Valid() bool {
	switch p {
	case DayParityEven, DayParityOdd:
		return true
	default:
		return false
	}
}
