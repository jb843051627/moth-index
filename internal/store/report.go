package store

import (
	"context"
	"fmt"

	"github.com/jb843051627/moth-index/internal/model"
)

type ReportStore struct{ db *DB }

func NewReportStore(db *DB) *ReportStore { return &ReportStore{db: db} }

func (s *ReportStore) BatchSummary(ctx context.Context, batchID int64) (model.BatchSummary, error) {
	var summary model.BatchSummary
	err := s.db.QueryRow(ctx, `SELECT id,station_id,trap_id,started_at,ended_at,status,weather_note FROM batches WHERE id=?`, batchID).
		Scan(&summary.Batch.ID, &summary.Batch.StationID, &summary.Batch.TrapID, &summary.Batch.StartedAt, &summary.Batch.EndedAt, &summary.Batch.Status, &summary.Batch.WeatherNote)
	if err != nil {
		if isNoRows(err) {
			return model.BatchSummary{}, ErrNotFound
		}
		return model.BatchSummary{}, fmt.Errorf("get summary batch: %w", err)
	}
	if err := s.db.QueryRow(ctx, `SELECT COALESCE(SUM(count),0) FROM specimens WHERE batch_id=?`, batchID).Scan(&summary.Specimens); err != nil {
		return model.BatchSummary{}, fmt.Errorf("summary specimens: %w", err)
	}
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM specimens WHERE batch_id=? AND status IN (?,?)`, batchID, model.SpecimenAccepted, model.SpecimenRejected).Scan(&summary.Classified); err != nil {
		return model.BatchSummary{}, fmt.Errorf("summary classifications: %w", err)
	}
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM reviews r JOIN specimens s ON s.id=r.specimen_id WHERE s.batch_id=? AND r.decision IN (?,?)`, batchID, model.ReviewApproved, model.ReviewRejected).Scan(&summary.Reviews); err != nil {
		return model.BatchSummary{}, fmt.Errorf("summary reviews: %w", err)
	}
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM signals WHERE batch_id=? AND active=1 AND severity=?`, batchID, model.SeverityBlocker).Scan(&summary.Blockers); err != nil {
		return model.BatchSummary{}, fmt.Errorf("summary blockers: %w", err)
	}
	if summary.Specimens == 0 {
		summary.Quality = 0
	} else {
		summary.Quality = float64(summary.Classified) / float64(summary.Specimens)
	}
	reading, err := s.latestReading(ctx, batchID)
	if err == nil {
		summary.LastReading = reading
	} else if err != ErrNotFound {
		return model.BatchSummary{}, err
	}
	return summary, nil
}

func (s *ReportStore) latestReading(ctx context.Context, batchID int64) (*model.Reading, error) {
	var reading model.Reading
	err := s.db.QueryRow(ctx, `SELECT id,batch_id,observed_at,temperature,humidity,lux,rainfall,source FROM readings WHERE batch_id=? ORDER BY observed_at DESC,id DESC LIMIT 1`, batchID).
		Scan(&reading.ID, &reading.BatchID, &reading.ObservedAt, &reading.Temperature, &reading.Humidity, &reading.Lux, &reading.Rainfall, &reading.Source)
	if isNoRows(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("summary latest reading: %w", err)
	}
	return &reading, nil
}

func (s *ReportStore) Daily(ctx context.Context, from, to string, limit int) ([]model.DailyRecord, error) {
	if limit <= 0 || limit > 366 {
		limit = 31
	}
	query := `SELECT substr(b.started_at,1,10),COUNT(DISTINCT b.id),COALESCE((SELECT SUM(s2.count) FROM specimens s2 JOIN batches b2 ON b2.id=s2.batch_id WHERE substr(b2.started_at,1,10)=substr(b.started_at,1,10)),0),COALESCE((SELECT COUNT(s3.id) FROM specimens s3 JOIN batches b3 ON b3.id=s3.batch_id WHERE substr(b3.started_at,1,10)=substr(b.started_at,1,10)),0),COALESCE((SELECT AVG(r2.humidity) FROM readings r2 JOIN batches b4 ON b4.id=r2.batch_id WHERE substr(b4.started_at,1,10)=substr(b.started_at,1,10)),0) FROM batches b WHERE 1=1`
	args := make([]any, 0, 3)
	if from != "" {
		query += ` AND b.started_at>=?`
		args = append(args, from)
	}
	if to != "" {
		query += ` AND b.started_at<=?`
		args = append(args, to)
	}
	query += ` GROUP BY substr(b.started_at,1,10) ORDER BY substr(b.started_at,1,10) DESC LIMIT ?`
	args = append(args, limit)
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("daily report: %w", err)
	}
	values := make([]model.DailyRecord, 0)
	for rows.Next() {
		var record model.DailyRecord
		if err := rows.Scan(&record.Day, &record.Batches, &record.Specimens, &record.Captures, &record.Quality); err != nil {
			return nil, closeRows(rows, fmt.Errorf("scan daily report: %w", err))
		}
		values = append(values, record)
	}
	return values, closeRows(rows, nil)
}

func (s *ReportStore) Audit(ctx context.Context, entity string, entityID int64, action, detail string) error {
	_, err := s.db.Exec(ctx, `INSERT INTO audit_events(entity,entity_id,action,detail,created_at) VALUES(?,?,?,?,?)`, entity, entityID, action, detail, nowText())
	if err != nil {
		return fmt.Errorf("audit event: %w", err)
	}
	return nil
}

func isNoRows(err error) bool { return err != nil && err.Error() == "sql: no rows in result set" }
