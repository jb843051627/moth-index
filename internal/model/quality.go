package model

import (
	"fmt"
	"sort"
	"strings"
)

type QualityFinding struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Detail   string `json:"detail"`
}

func (f QualityFinding) Blocks() bool { return f.Severity == SeverityBlocker }

func FindingsForReadings(values []Reading) []QualityFinding {
	findings := make([]QualityFinding, 0)
	if len(values) == 0 {
		return append(findings, QualityFinding{Code: "no-readings", Severity: SeverityBlocker, Detail: "batch has no environment readings"})
	}
	for _, value := range values {
		if value.Temperature < -5 || value.Temperature > 45 {
			findings = append(findings, QualityFinding{Code: "temperature-range", Severity: SeverityWarn, Detail: fmt.Sprintf("temperature %.1f is outside field range", value.Temperature)})
		}
		if value.Humidity > 98 {
			findings = append(findings, QualityFinding{Code: "humidity-saturation", Severity: SeverityWarn, Detail: "humidity sensor is saturated"})
		}
		if value.Rainfall > 20 {
			findings = append(findings, QualityFinding{Code: "rain-event", Severity: SeverityInfo, Detail: "heavy rain may bias light trap capture"})
		}
	}
	return findings
}

func FindingsForSpecimens(values []Specimen) []QualityFinding {
	findings := make([]QualityFinding, 0)
	seen := make(map[string]int)
	for _, value := range values {
		tag := strings.ToUpper(strings.TrimSpace(value.Tag))
		if tag == "" {
			findings = append(findings, QualityFinding{Code: "empty-tag", Severity: SeverityBlocker, Detail: "specimen tag is empty"})
		}
		if seen[tag] > 0 {
			findings = append(findings, QualityFinding{Code: "duplicate-tag", Severity: SeverityBlocker, Detail: "specimen tag appears more than once"})
		}
		seen[tag]++
		if value.Count > 1000 {
			findings = append(findings, QualityFinding{Code: "large-count", Severity: SeverityWarn, Detail: "count needs a second reading"})
		}
		if !value.IsClassified() {
			findings = append(findings, QualityFinding{Code: "unclassified", Severity: SeverityInfo, Detail: "specimen still needs taxonomic review"})
		}
	}
	sort.SliceStable(findings, func(i, j int) bool { return SeverityRank(findings[i].Severity) > SeverityRank(findings[j].Severity) })
	return findings
}

func HighestFinding(findings []QualityFinding) QualityFinding {
	if len(findings) == 0 {
		return QualityFinding{}
	}
	highest := findings[0]
	for _, finding := range findings[1:] {
		if SeverityRank(finding.Severity) > SeverityRank(highest.Severity) {
			highest = finding
		}
	}
	return highest
}

func HasBlockingFinding(findings []QualityFinding) bool {
	for _, finding := range findings {
		if finding.Blocks() {
			return true
		}
	}
	return false
}
