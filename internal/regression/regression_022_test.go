package regression

import (
	"testing"

	"github.com/jb843051627/moth-index/internal/model"
)

func TestBug22_EmptyTemperatureStatsAreSafe(t *testing.T) {
	if got := model.MedianTemperature(nil); got != 0 { t.Fatalf("empty median=%v", got) }
	if got := model.ReadingVariance([]model.Reading{{Temperature: 20}}); got != 0 { t.Fatalf("single variance=%v", got) }
}
