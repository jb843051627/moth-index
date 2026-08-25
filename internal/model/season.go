package model

import (
	"math"
	"sort"
	"time"
)

type Season string

const (
	SeasonSpring Season = "spring"
	SeasonSummer Season = "summer"
	SeasonAutumn Season = "autumn"
	SeasonWinter Season = "winter"
)

func SeasonOf(month time.Month) Season {
	switch month {
	case time.March, time.April, time.May:
		return SeasonSpring
	case time.June, time.July, time.August:
		return SeasonSummer
	case time.September, time.October, time.November:
		return SeasonAutumn
	default:
		return SeasonWinter
	}
}

type SeasonalPulse struct {
	Season      Season  `json:"season"`
	Nights      int     `json:"nights"`
	Individuals int     `json:"individuals"`
	AverageRate float64 `json:"average_rate"`
	RainNights  int     `json:"rain_nights"`
}

func BuildSeasonalPulses(batches []NightBatch, specimens map[int64][]Specimen, readings map[int64][]Reading) []SeasonalPulse {
	grouped := make(map[Season]*SeasonalPulse)
	for _, batch := range batches {
		stamp, err := time.Parse(time.RFC3339Nano, batch.StartedAt)
		if err != nil {
			continue
		}
		season := SeasonOf(stamp.Month())
		pulse := grouped[season]
		if pulse == nil {
			pulse = &SeasonalPulse{Season: season}
			grouped[season] = pulse
		}
		pulse.Nights++
		total := 0
		for _, specimen := range specimens[batch.ID] {
			total += specimen.Count
		}
		pulse.Individuals += total
		values := readings[batch.ID]
		for _, reading := range values {
			if reading.IsWet() {
				pulse.RainNights++
				break
			}
		}
	}
	output := make([]SeasonalPulse, 0, len(grouped))
	for _, pulse := range grouped {
		if pulse.Nights > 0 {
			pulse.AverageRate = float64(pulse.Individuals) / float64(pulse.Nights)
		}
		output = append(output, *pulse)
	}
	sort.Slice(output, func(i, j int) bool { return output[i].Season < output[j].Season })
	return output
}

func MoonIllumination(day time.Time) float64 {
	days := float64(day.Unix()/86400 - 4)
	phase := math.Mod(days, 29.53059) / 29.53059
	if phase < 0 {
		phase += 1
	}
	return (1 - math.Cos(phase*2*math.Pi)) / 2
}

func MoonSuppressed(day time.Time) bool { return MoonIllumination(day) > 0.2 }

func EmergenceScore(readings []Reading, day time.Time) float64 {
	if len(readings) == 0 || MoonSuppressed(day) {
		return 0
	}
	score := 0.0
	for _, reading := range readings {
		if reading.IsWarm() {
			score += 1
		}
		if reading.Humidity >= 60 && reading.Humidity <= 90 {
			score += 0.5
		}
		if reading.IsWet() {
			score -= 0.75
		}
	}
	score /= float64(len(readings))
	if score < 0 {
		return 0
	}
	if score > 1 {
		return 1
	}
	return score
}

func IsSurveyNight(start time.Time) bool { return start.Hour() >= 18 || start.Hour() <= 4 }

func NormalizeSeason(value string) Season {
	switch value {
	case string(SeasonSpring):
		return SeasonSpring
	case string(SeasonSummer):
		return SeasonSummer
	case string(SeasonAutumn):
		return SeasonAutumn
	default:
		return SeasonWinter
	}
}
