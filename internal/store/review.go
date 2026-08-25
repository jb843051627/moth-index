package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jb843051627/moth-index/internal/model"
)

type ReviewStore struct{ db *DB }

func NewReviewStore(db *DB) *ReviewStore { return &ReviewStore{db: db} }

func (s *ReviewStore) Create(ctx context.Context, review model.Review) (model.Review, error) {
	result, err := s.db.Exec(ctx, `INSERT INTO reviews(specimen_id,reviewer,decision,confidence,note,reviewed_at) VALUES(?,?,?,?,?,?)`,
		review.SpecimenID, review.Reviewer, review.Decision, review.Confidence, review.Note, review.ReviewedAt)
	if err != nil {
		return model.Review{}, fmt.Errorf("insert review: %w", err)
	}
	review.ID, err = txInsertID(result)
	if err != nil {
		return model.Review{}, err
	}
	return review, nil
}

func (s *ReviewStore) LatestBySpecimen(ctx context.Context, specimenID int64) (model.Review, error) {
	var review model.Review
	err := s.db.QueryRow(ctx, `SELECT id,specimen_id,reviewer,decision,confidence,note,reviewed_at FROM reviews WHERE specimen_id=? ORDER BY reviewed_at DESC,id DESC LIMIT 1`, specimenID).
		Scan(&review.ID, &review.SpecimenID, &review.Reviewer, &review.Decision, &review.Confidence, &review.Note, &review.ReviewedAt)
	if err == sql.ErrNoRows {
		return model.Review{}, ErrNotFound
	}
	if err != nil {
		return model.Review{}, fmt.Errorf("get latest review: %w", err)
	}
	return review, nil
}

func (s *ReviewStore) ListByBatch(ctx context.Context, batchID int64, limit int) ([]model.Review, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := s.db.Query(ctx, `SELECT r.id,r.specimen_id,r.reviewer,r.decision,r.confidence,r.note,r.reviewed_at FROM reviews r JOIN specimens s ON s.id=r.specimen_id WHERE s.batch_id=? ORDER BY r.reviewed_at DESC,r.id DESC LIMIT ?`, batchID, limit)
	if err != nil {
		return nil, fmt.Errorf("list reviews: %w", err)
	}
	values := make([]model.Review, 0)
	for rows.Next() {
		var review model.Review
		if err := rows.Scan(&review.ID, &review.SpecimenID, &review.Reviewer, &review.Decision, &review.Confidence, &review.Note, &review.ReviewedAt); err != nil {
			return nil, closeRows(rows, fmt.Errorf("scan review: %w", err))
		}
		values = append(values, review)
	}
	return values, closeRows(rows, nil)
}

func (s *ReviewStore) CountFinalByBatch(ctx context.Context, batchID int64) (int, error) {
	var count int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM reviews r JOIN specimens s ON s.id=r.specimen_id WHERE s.batch_id=? AND r.decision IN (?,?)`, batchID, model.ReviewApproved, model.ReviewRejected).Scan(&count); err != nil {
		return 0, fmt.Errorf("count final reviews: %w", err)
	}
	return count, nil
}
