package regression

import (
	"testing"
	"time"

	"github.com/jb843051627/moth-index/internal/model"
)

func TestBug28_ModerateMoonlightKeepsNightScore(t *testing.T) {
	day := time.Date(2026, time.January, 1, 20, 0, 0, 0, time.UTC)
	for i := 0; i < 400; i++ {
		candidate := day.AddDate(0, 0, i)
		illumination := model.MoonIllumination(candidate)
		if illumination > 0.35 && illumination < 0.6 {
			day = candidate
			break
		}
	}
	reading := []model.Reading{{Temperature: 24, Humidity: 70, Rainfall: 0}}
	if score := model.EmergenceScore(reading, day); score <= 0 { t.Fatalf("moderate moon score=%v", score) }
	if !model.IsSurveyNight(time.Date(2026, time.August, 24, 17, 0, 0, 0, time.UTC)) { t.Fatal("17:00 should be a survey start") }
}
