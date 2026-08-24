package service

import (
	"context"
	"fmt"

	"github.com/jb843051627/moth-index/internal/model"
	"github.com/jb843051627/moth-index/internal/store"
)

type ReviewService struct {
	reviews   *store.ReviewStore
	specimens *store.SpecimenStore
	reports   *store.ReportStore
	tasks     *store.TaskStore
}

func NewReviewService(reviews *store.ReviewStore, specimens *store.SpecimenStore, reports *store.ReportStore, tasks *store.TaskStore) *ReviewService {
	return &ReviewService{reviews: reviews, specimens: specimens, reports: reports, tasks: tasks}
}

func (s *ReviewService) Submit(ctx context.Context, input model.Review) (model.Review, error) {
	input.ReviewedAt = serviceNow()
	if input.Decision == "" {
		input.Decision = model.ReviewPending
	}
	if err := input.Validate(); err != nil {
		return model.Review{}, wrapValidation(err)
	}
	specimen, err := s.specimens.Get(ctx, input.SpecimenID)
	if err != nil {
		return model.Review{}, err
	}
	if !specimen.CanReview() {
		return model.Review{}, fmt.Errorf("%w: specimen cannot be reviewed", store.ErrState)
	}
	review, err := s.reviews.Create(ctx, input)
	if err != nil {
		return model.Review{}, err
	}
	status := model.SpecimenAccepted
	if input.Decision == model.ReviewApproved {
		status = model.SpecimenAccepted
	}
	if input.Decision == model.ReviewRejected {
		status = model.SpecimenRejected
	}
	if input.IsFinal() {
		if err := s.specimens.UpdateClassification(ctx, specimen.ID, specimen.Family, specimen.Genus, specimen.Species, status); err != nil {
			return model.Review{}, err
		}
	}
	return review, s.reports.Audit(ctx, "specimen", specimen.ID, "review", input.Decision)
}

func (s *ReviewService) Approve(ctx context.Context, specimenID int64, reviewer string, confidence float64) (model.Review, error) {
	return s.Submit(ctx, model.Review{SpecimenID: specimenID, Reviewer: reviewer, Decision: model.ReviewApproved, Confidence: confidence})
}

func (s *ReviewService) Reject(ctx context.Context, specimenID int64, reviewer, note string) (model.Review, error) {
	return s.Submit(ctx, model.Review{SpecimenID: specimenID, Reviewer: reviewer, Decision: model.ReviewRejected, Confidence: 0, Note: note})
}

func (s *ReviewService) Latest(ctx context.Context, specimenID int64) (model.Review, error) {
	return s.reviews.LatestBySpecimen(ctx, specimenID)
}

func (s *ReviewService) ListForBatch(ctx context.Context, batchID int64, limit int) ([]model.Review, error) {
	return s.reviews.ListByBatch(ctx, batchID, limit)
}

func (s *ReviewService) Process(ctx context.Context, task model.ReviewTask) error {
	specimen, err := s.specimens.Get(ctx, task.RefID)
	if err != nil {
		return err
	}
	if !specimen.IsClassified() {
		return fmt.Errorf("%w: specimen is not classified", store.ErrState)
	}
	_, err = s.Approve(ctx, specimen.ID, "queue-worker", 0.7)
	return err
}

func (s *ReviewService) FinalCount(ctx context.Context, batchID int64) (int, error) {
	return s.reviews.CountFinalByBatch(ctx, batchID)
}
