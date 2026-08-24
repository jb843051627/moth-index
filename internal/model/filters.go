package model

import (
	"strings"
	"time"
)

type BatchFilter struct {
	StationID int64
	Status    string
	From      string
	To        string
}

func (f BatchFilter) HasWindow() bool { return f.From != "" || f.To != "" }

func (f BatchFilter) ValidWindow() bool {
	if !f.HasWindow() {
		return true
	}
	from, e1 := time.Parse(time.RFC3339Nano, f.From)
	to, e2 := time.Parse(time.RFC3339Nano, f.To)
	return e1 == nil && e2 == nil && !to.Before(from)
}

func CleanStatus(value string) string { return strings.ToLower(strings.TrimSpace(value)) }

func ParsePage(limit, offset int) Page { return (Page{Limit: limit, Offset: offset}).Normalize() }
