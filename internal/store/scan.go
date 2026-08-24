package store

import (
	"database/sql"
	"fmt"
	"time"
)

func nowText() string { return time.Now().UTC().Format(time.RFC3339Nano) }

func scanString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func closeRows(rows *sql.Rows, current error) error {
	if err := rows.Close(); current == nil && err != nil {
		return fmt.Errorf("close rows: %w", err)
	}
	if err := rows.Err(); current == nil && err != nil {
		return fmt.Errorf("iterate rows: %w", err)
	}
	return current
}

func pageValues(limit, offset int) (int, int) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
