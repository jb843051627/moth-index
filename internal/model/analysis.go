package model

import (
	"math"
	"sort"
	"strings"
	"time"
)

type CaptureProfile struct {
	TotalIndividuals int
	TaxaCount        int
	FemaleCount      int
	MaleCount        int
	UnknownSex       int
	DominantFamily   string
	Diversity        float64
}

func BuildCaptureProfile(values []Specimen) CaptureProfile {
	profile := CaptureProfile{}
	families := make(map[string]int)
	for _, value := range values {
		profile.TotalIndividuals += value.Count
		if value.IsClassified() {
			profile.TaxaCount++
			families[value.Family] += value.Count
		}
		switch strings.ToLower(value.Sex) {
		case "female":
			profile.FemaleCount += value.Count
		case "male":
			profile.MaleCount += value.Count
		default:
			profile.UnknownSex += value.Count
		}
	}
	for family, count := range families {
		if count > families[profile.DominantFamily] {
			profile.DominantFamily = family
		}
	}
	if profile.TotalIndividuals > 0 && len(families) > 0 {
		profile.Diversity = float64(len(families)) / float64(profile.TotalIndividuals)
	}
	return profile
}

type Window struct {
	Start time.Time
	End   time.Time
}

func (w Window) Valid() bool { return !w.Start.IsZero() && !w.End.IsZero() && !w.End.Before(w.Start) }

func (w Window) Contains(value time.Time) bool {
	return w.Valid() && !value.Before(w.Start) && !value.After(w.End)
}

func (w Window) Duration() time.Duration {
	if !w.Valid() {
		return 0
	}
	return w.End.Sub(w.Start)
}

func (w Window) NightHours() float64 { return w.Duration().Hours() }

func ParseWindow(start, end string) (Window, error) {
	left, err := time.Parse(time.RFC3339Nano, start)
	if err != nil {
		return Window{}, err
	}
	right, err := time.Parse(time.RFC3339Nano, end)
	if err != nil {
		return Window{}, err
	}
	return Window{Start: left, End: right}, nil
}

type ObservationBand string

const (
	BandCold  ObservationBand = "cold"
	BandMild  ObservationBand = "mild"
	BandWarm  ObservationBand = "warm"
	BandStorm ObservationBand = "storm"
)

func BandOf(reading Reading) ObservationBand {
	if reading.Rainfall >= 8 || reading.Humidity >= 98 {
		return BandStorm
	}
	if reading.Temperature < 10 {
		return BandCold
	}
	if reading.Temperature >= 22 && reading.Humidity >= 55 {
		return BandWarm
	}
	return BandMild
}

func RankBands(values []Reading) []ObservationBand {
	counts := map[ObservationBand]int{}
	for _, value := range values {
		counts[BandOf(value)]++
	}
	output := make([]ObservationBand, 0, len(counts))
	for band := range counts {
		output = append(output, band)
	}
	sort.Slice(output, func(i, j int) bool { return counts[output[i]] > counts[output[j]] })
	return output
}

func MedianTemperature(values []Reading) float64 {
	if len(values) == 0 {
		return 0
	}
	temperatures := make([]float64, 0, len(values))
	for _, value := range values {
		temperatures = append(temperatures, value.Temperature)
	}
	sort.Float64s(temperatures)
	middle := len(temperatures) / 2
	if len(temperatures)%2 == 0 {
		return (temperatures[middle-1] + temperatures[middle]) / 2
	}
	return temperatures[middle]
}

func ReadingVariance(values []Reading) float64 {
	if len(values) < 2 {
		return 0
	}
	mean := 0.0
	for _, value := range values {
		mean += value.Temperature
	}
	mean /= float64(len(values))
	total := 0.0
	for _, value := range values {
		delta := value.Temperature - mean
		total += delta * delta
	}
	variance := total / float64(len(values)-1)
	if math.IsNaN(variance) || math.IsInf(variance, 0) {
		return 0
	}
	return variance
}
