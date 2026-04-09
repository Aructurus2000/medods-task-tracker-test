package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

const taskColumns = `id, title, description, status, recurrence_type, recurrence_every_n_days,
	recurrence_day_of_month, recurrence_specific_dates, recurrence_parity, created_at, updated_at`

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	query := fmt.Sprintf(`
		INSERT INTO tasks (
			title, description, status, recurrence_type, recurrence_every_n_days,
			recurrence_day_of_month, recurrence_specific_dates, recurrence_parity, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9, $10)
		RETURNING %s
	`, taskColumns)

	recurrenceDates, err := marshalSpecificDates(task.Recurrence.SpecificDates)
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Status,
		task.Recurrence.Type,
		nullableInt(task.Recurrence.EveryNDays),
		nullableInt(task.Recurrence.DayOfMonth),
		recurrenceDates,
		nullableString(string(task.Recurrence.Parity)),
		task.CreatedAt,
		task.UpdatedAt,
	)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM tasks
		WHERE id = $1
	`, taskColumns)

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return found, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	query := fmt.Sprintf(`
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			recurrence_type = $4,
			recurrence_every_n_days = $5,
			recurrence_day_of_month = $6,
			recurrence_specific_dates = $7::jsonb,
			recurrence_parity = $8,
			updated_at = $9
		WHERE id = $10
		RETURNING %s
	`, taskColumns)

	recurrenceDates, err := marshalSpecificDates(task.Recurrence.SpecificDates)
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Status,
		task.Recurrence.Type,
		nullableInt(task.Recurrence.EveryNDays),
		nullableInt(task.Recurrence.DayOfMonth),
		recurrenceDates,
		nullableString(string(task.Recurrence.Parity)),
		task.UpdatedAt,
		task.ID,
	)
	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context, onDate *time.Time) ([]taskdomain.Task, error) {
	query := buildListQuery(onDate)
	args := make([]any, 0)
	if onDate != nil {
		args = append(args, onDate.Format("2006-01-02"))
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *Repository) ListOpenDueOnDate(ctx context.Context, onDate time.Time) ([]taskdomain.Task, error) {
	query := buildListOpenDueQuery(onDate)
	rows, err := r.pool.Query(ctx, query, onDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task                    taskdomain.Task
		status                  string
		recurrenceType          string
		recurrenceEveryNDays    *int
		recurrenceDayOfMonth    *int
		recurrenceSpecificDates []byte
		recurrenceParity        *string
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&recurrenceType,
		&recurrenceEveryNDays,
		&recurrenceDayOfMonth,
		&recurrenceSpecificDates,
		&recurrenceParity,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)
	task.Recurrence.Type = taskdomain.RecurrenceType(recurrenceType)
	if recurrenceEveryNDays != nil {
		task.Recurrence.EveryNDays = *recurrenceEveryNDays
	}
	if recurrenceDayOfMonth != nil {
		task.Recurrence.DayOfMonth = *recurrenceDayOfMonth
	}
	if recurrenceParity != nil {
		task.Recurrence.Parity = taskdomain.DayParity(*recurrenceParity)
	}
	if len(recurrenceSpecificDates) > 0 {
		var rawDates []string
		if err := json.Unmarshal(recurrenceSpecificDates, &rawDates); err != nil {
			return nil, fmt.Errorf("decode recurrence_specific_dates: %w", err)
		}
		task.Recurrence.SpecificDates = make([]time.Time, 0, len(rawDates))
		for _, rawDate := range rawDates {
			date, err := time.Parse("2006-01-02", rawDate)
			if err != nil {
				return nil, fmt.Errorf("decode recurrence_specific_dates date: %w", err)
			}
			task.Recurrence.SpecificDates = append(task.Recurrence.SpecificDates, date.UTC())
		}
	}

	return &task, nil
}

func marshalSpecificDates(dates []time.Time) ([]byte, error) {
	if len(dates) == 0 {
		return []byte("[]"), nil
	}
	values := make([]string, 0, len(dates))
	for _, d := range dates {
		values = append(values, d.UTC().Format("2006-01-02"))
	}
	result, err := json.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("encode recurrence_specific_dates: %w", err)
	}
	return result, nil
}

func nullableInt(v int) any {
	if v <= 0 {
		return nil
	}
	return v
}

func nullableString(v string) any {
	if v == "" {
		return nil
	}
	return v
}

func buildListQuery(onDate *time.Time) string {
	query := fmt.Sprintf(`
		SELECT %s
		FROM tasks
	`, taskColumns)
	if onDate != nil {
		query += `
		WHERE (
			recurrence_type = 'none'
			OR (
				recurrence_type = 'every_n_days'
				AND $1::date >= created_at::date
				AND ($1::date - created_at::date) % recurrence_every_n_days = 0
			)
			OR (recurrence_type = 'monthly' AND EXTRACT(day FROM $1::date) = recurrence_day_of_month)
			OR (recurrence_type = 'specific_dates' AND recurrence_specific_dates @> to_jsonb(ARRAY[to_char($1::date, 'YYYY-MM-DD')]))
			OR (recurrence_type = 'month_day_parity' AND (
				(recurrence_parity = 'even' AND MOD(EXTRACT(day FROM $1::date)::int, 2) = 0)
				OR (recurrence_parity = 'odd' AND MOD(EXTRACT(day FROM $1::date)::int, 2) = 1)
			))
		)
	`
	}
	query += `
		ORDER BY id DESC
	`
	return query
}

func buildListOpenDueQuery(onDate time.Time) string {
	query := fmt.Sprintf(`
		SELECT %s
		FROM tasks
	`, taskColumns)
	query += `
		WHERE (
			recurrence_type = 'none'
			OR (
				recurrence_type = 'every_n_days'
				AND $1::date >= created_at::date
				AND ($1::date - created_at::date) % recurrence_every_n_days = 0
			)
			OR (recurrence_type = 'monthly' AND EXTRACT(day FROM $1::date) = recurrence_day_of_month)
			OR (recurrence_type = 'specific_dates' AND recurrence_specific_dates @> to_jsonb(ARRAY[to_char($1::date, 'YYYY-MM-DD')]))
			OR (recurrence_type = 'month_day_parity' AND (
				(recurrence_parity = 'even' AND MOD(EXTRACT(day FROM $1::date)::int, 2) = 0)
				OR (recurrence_parity = 'odd' AND MOD(EXTRACT(day FROM $1::date)::int, 2) = 1)
			))
		)
		AND status != 'done'
	`
	query += `
		ORDER BY id DESC
	`
	return query
}
