package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		Recurrence:  normalized.Recurrence,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		Recurrence:  normalized.Recurrence,
		UpdatedAt:   s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context, onDate *time.Time) ([]taskdomain.Task, error) {
	return s.repo.List(ctx, onDate)
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	normalizedRecurrence, err := validateRecurrence(input.Recurrence)
	if err != nil {
		return CreateInput{}, err
	}
	input.Recurrence = normalizedRecurrence

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	normalizedRecurrence, err := validateRecurrence(input.Recurrence)
	if err != nil {
		return UpdateInput{}, err
	}
	input.Recurrence = normalizedRecurrence

	return input, nil
}

func validateRecurrence(input taskdomain.Recurrence) (taskdomain.Recurrence, error) {
	if input.Type == "" {
		input.Type = taskdomain.RecurrenceNone
	}
	if !input.Type.Valid() {
		return taskdomain.Recurrence{}, fmt.Errorf("%w: invalid recurrence type", ErrInvalidInput)
	}

	switch input.Type {
	case taskdomain.RecurrenceNone:
		return taskdomain.Recurrence{Type: taskdomain.RecurrenceNone}, nil
	case taskdomain.RecurrenceEveryNDay:
		if input.EveryNDays <= 0 {
			return taskdomain.Recurrence{}, fmt.Errorf("%w: every_n_days must be positive", ErrInvalidInput)
		}
		return taskdomain.Recurrence{
			Type:       input.Type,
			EveryNDays: input.EveryNDays,
		}, nil
	case taskdomain.RecurrenceMonthly:
		if input.DayOfMonth < 1 || input.DayOfMonth > 31 {
			return taskdomain.Recurrence{}, fmt.Errorf("%w: day_of_month must be in range 1..31", ErrInvalidInput)
		}
		return taskdomain.Recurrence{
			Type:       input.Type,
			DayOfMonth: input.DayOfMonth,
		}, nil
	case taskdomain.RecurrenceDates:
		if len(input.SpecificDates) == 0 {
			return taskdomain.Recurrence{}, fmt.Errorf("%w: specific_dates must not be empty", ErrInvalidInput)
		}
		normalizedDates := make([]time.Time, 0, len(input.SpecificDates))
		seen := make(map[string]struct{}, len(input.SpecificDates))
		for _, d := range input.SpecificDates {
			day := time.Date(d.UTC().Year(), d.UTC().Month(), d.UTC().Day(), 0, 0, 0, 0, time.UTC)
			key := day.Format("2006-01-02")
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			normalizedDates = append(normalizedDates, day)
		}
		return taskdomain.Recurrence{
			Type:          input.Type,
			SpecificDates: normalizedDates,
		}, nil
	case taskdomain.RecurrenceParity:
		if !input.Parity.Valid() {
			return taskdomain.Recurrence{}, fmt.Errorf("%w: parity must be even or odd", ErrInvalidInput)
		}
		return taskdomain.Recurrence{
			Type:   input.Type,
			Parity: input.Parity,
		}, nil
	default:
		return taskdomain.Recurrence{}, fmt.Errorf("%w: invalid recurrence type", ErrInvalidInput)
	}
}
