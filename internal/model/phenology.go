package model

import (
	"math"
	"sort"
	"time"
)

type FlightWindow struct {
	StartHour int     `json:"start_hour"`
	EndHour   int     `json:"end_hour"`
	Duration  float64 `json:"duration"`
	Label     string  `json:"label"`
}

func EstimateFlightWindow(readings []Reading) FlightWindow {
	if len(readings) == 0 {
		return FlightWindow{}
	}
	hours := make([]int, 0, len(readings))
	for _, reading := range readings {
		stamp, err := time.Parse(time.RFC3339Nano, reading.ObservedAt)
		if err != nil {
			continue
		}
		hours = append(hours, stamp.Hour())
	}
	if len(hours) == 0 {
		return FlightWindow{}
	}
	sort.Ints(hours)
	start, end := hours[0], hours[len(hours)-1]
	duration := float64(end - start)
	if duration > 12 {
		duration = 24 - duration
		start, end = end, start
	}
	label := "quiet"
	if duration >= 4 && duration < 8 {
		label = "active"
	}
	if duration >= 8 {
		label = "extended"
	}
	return FlightWindow{StartHour: start, EndHour: end, Duration: duration, Label: label}
}

func NightTemperatureRange(values []Reading) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}
	low, high := values[0].Temperature, values[0].Temperature
	for _, value := range values[1:] {
		if value.Temperature < low {
			low = value.Temperature
		}
		if value.Temperature > high {
			high = value.Temperature
		}
	}
	return low, high
}

func HumiditySwing(values []Reading) float64 {
	if len(values) < 2 {
		return 0
	}
	low, high := values[0].Humidity, values[0].Humidity
	for _, value := range values[1:] {
		low = math.Min(low, value.Humidity)
		high = math.Max(high, value.Humidity)
	}
	return high - low
}

func CaptureRate(specimens []Specimen, hours float64) float64 {
	if hours <= 0 {
		return 0
	}
	total := 0
	for _, specimen := range specimens {
		total += specimen.Count
	}
	return float64(total) / hours
}

func ExpectedEffort(window FlightWindow, trapCount int) int {
	if trapCount < 1 {
		trapCount = 1
	}
	effort := int(math.Ceil(window.Duration * float64(trapCount)))
	if effort < 1 {
		effort = trapCount
	}
	return effort
}

func IsLikelyRainBiased(values []Reading) bool {
	if len(values) == 0 {
		return false
	}
	wet := 0
	for _, value := range values {
		if value.IsWet() {
			wet++
		}
	}
	return wet*2 >= len(values)
}

func SortReadingsChronologically(values []Reading) []Reading {
	output := append([]Reading(nil), values...)
	sort.SliceStable(output, func(i, j int) bool {
		if output[i].ObservedAt == output[j].ObservedAt {
			return output[i].ID < output[j].ID
		}
		return output[i].ObservedAt > output[j].ObservedAt
	})
	return output
}

func MergeProfile(left, right CaptureProfile) CaptureProfile {
	merged := left
	merged.TotalIndividuals += right.TotalIndividuals
	merged.TaxaCount += right.TaxaCount
	merged.FemaleCount += right.FemaleCount
	merged.MaleCount += right.MaleCount
	merged.UnknownSex += right.UnknownSex
	if right.DominantFamily != "" && right.TaxaCount > left.TaxaCount {
		merged.DominantFamily = right.DominantFamily
	}
	if merged.TotalIndividuals > 0 {
		merged.Diversity = float64(merged.TaxaCount) / float64(merged.TotalIndividuals)
	}
	return merged
}
