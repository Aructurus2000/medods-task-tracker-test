package postgres

import (
	"strings"
	"testing"
	"time"
)

func TestListQuery_ContainsGuardForDatesBeforeCreatedAt(t *testing.T) {
	query := buildListQuery(datePtr(time.Date(2026, time.April, 10, 0, 0, 0, 0, time.UTC)))

	if !strings.Contains(query, "AND $1::date >= created_at::date") {
		t.Fatalf("every_n_days guard is missing in query: %s", query)
	}
}

func datePtr(v time.Time) *time.Time {
	return &v
}
