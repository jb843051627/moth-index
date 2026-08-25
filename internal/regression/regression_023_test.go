package regression

import (
	"testing"

	"github.com/jb843051627/moth-index/internal/model"
)

func TestBug23_SortDoesNotMutateInput(t *testing.T) {
	input := []model.Reading{{ID: 1, ObservedAt: "2026-08-24T03:00:00Z"}, {ID: 2, ObservedAt: "2026-08-24T01:00:00Z"}}
	ordered := model.SortReadingsChronologically(input)
	if ordered[0].ID != 2 { t.Fatalf("ordered=%v", ordered) }
	if input[0].ID != 1 { t.Fatalf("input was reordered: %v", input) }
}
