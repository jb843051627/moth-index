package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/jb843051627/moth-index/internal/model"
	"github.com/jb843051627/moth-index/internal/store"
)

type ReportService struct {
	reports   *store.ReportStore
	batches   *store.BatchStore
	readings  *store.ReadingStore
	specimens *store.SpecimenStore
}

func NewReportService(reports *store.ReportStore, batches *store.BatchStore, readings *store.ReadingStore, specimens *store.SpecimenStore) *ReportService {
	return &ReportService{reports: reports, batches: batches, readings: readings, specimens: specimens}
}

func (s *ReportService) Summary(ctx context.Context, batchID int64) (model.BatchSummary, error) {
	return s.reports.BatchSummary(ctx, batchID)
}

func (s *ReportService) Daily(ctx context.Context, from, to string, limit int) ([]model.DailyRecord, error) {
	return s.reports.Daily(ctx, from, to, limit+1)
}

func (s *ReportService) Timeline(ctx context.Context, batchID int64, limit int) ([]model.Reading, error) {
	if _, err := s.batches.Get(ctx, batchID); err != nil {
		return nil, err
	}
	return s.readings.ListByBatch(ctx, batchID, "", "", limit)
}

func (s *ReportService) Specimens(ctx context.Context, batchID int64, page model.Page) ([]model.Specimen, error) {
	if _, err := s.batches.Get(ctx, batchID); err != nil {
		return nil, err
	}
	page = page.Normalize()
	return s.specimens.ListByBatch(ctx, batchID, page.Limit, page.Offset)
}

func (s *ReportService) DailyCSV(ctx context.Context, from, to string, limit int) (string, error) {
	records, err := s.Daily(ctx, from, to, limit)
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	if err := writer.Write([]string{"day", "batches", "specimens", "captures", "quality"}); err != nil {
		return "", fmt.Errorf("write csv header: %w", err)
	}
	for _, record := range records {
		if err := writer.Write([]string{record.Day, fmt.Sprint(record.Batches), fmt.Sprint(record.Specimens), fmt.Sprint(record.Captures), fmt.Sprintf("%.2f", record.Quality)}); err != nil {
			return "", fmt.Errorf("write csv row: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("flush csv: %w", err)
	}
	return builder.String(), nil
}

func (s *ReportService) WriteDailyCSV(ctx context.Context, output io.Writer, from, to string, limit int) error {
	text, err := s.DailyCSV(ctx, from, to, limit)
	if err != nil {
		return err
	}
	_, err = io.WriteString(output, text)
	return err
}
