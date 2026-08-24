package service

import (
	"context"
	"fmt"

	"github.com/jb843051627/moth-index/internal/model"
	"github.com/jb843051627/moth-index/internal/store"
)

type SpecimenService struct {
	specimens *store.SpecimenStore
	batches   *store.BatchStore
	taxonomy  *store.TaxonomyStore
	tasks     *store.TaskStore
	reports   *store.ReportStore
}

func NewSpecimenService(specimens *store.SpecimenStore, batches *store.BatchStore, taxonomy *store.TaxonomyStore, tasks *store.TaskStore, reports *store.ReportStore) *SpecimenService {
	return &SpecimenService{specimens: specimens, batches: batches, taxonomy: taxonomy, tasks: tasks, reports: reports}
}

func (s *SpecimenService) Capture(ctx context.Context, input model.Specimen) (model.Specimen, error) {
	input.Tag = input.Tag
	input.Status = model.SpecimenCaptured
	input.CreatedAt = serviceNow()
	if err := input.Validate(); err != nil {
		return model.Specimen{}, wrapValidation(err)
	}
	batch, err := s.batches.Get(ctx, input.BatchID)
	if err != nil {
		return model.Specimen{}, err
	}
	if err := requireStatus(batch.Status, model.BatchOpen, model.BatchReview); err != nil {
		return model.Specimen{}, err
	}
	if _, err := s.specimens.GetByTag(ctx, input.Tag); err == nil {
		return model.Specimen{}, fmt.Errorf("%w: specimen tag", store.ErrConflict)
	} else if !isNotFound(err) {
		return model.Specimen{}, err
	}
	specimen, err := s.specimens.Create(ctx, input)
	if err != nil {
		return model.Specimen{}, err
	}
	task := model.ReviewTask{Kind: "specimen", RefID: specimen.ID, CreatedAt: serviceNow(), UpdatedAt: serviceNow()}
	if _, err := s.tasks.Enqueue(ctx, task); err != nil {
		return model.Specimen{}, err
	}
	return specimen, s.reports.Audit(ctx, "specimen", specimen.ID, "capture", specimen.Tag)
}

func (s *SpecimenService) Get(ctx context.Context, id int64) (model.Specimen, error) {
	return s.specimens.Get(ctx, id)
}

func (s *SpecimenService) List(ctx context.Context, batchID int64, page model.Page) ([]model.Specimen, error) {
	if _, err := s.batches.Get(ctx, batchID); err != nil {
		return nil, err
	}
	page = page.Normalize()
	return s.specimens.ListByBatch(ctx, batchID, page.Limit, page.Offset)
}

func (s *SpecimenService) Classify(ctx context.Context, id int64, taxon model.Taxon) (model.Specimen, error) {
	specimen, err := s.specimens.Get(ctx, id)
	if err != nil {
		return model.Specimen{}, err
	}
	if err := taxon.Validate(); err != nil {
		return model.Specimen{}, wrapValidation(err)
	}
	if _, err := s.taxonomy.Upsert(ctx, taxon); err != nil {
		return model.Specimen{}, err
	}
	if err := s.specimens.UpdateClassification(ctx, id, taxon.Family, taxon.Genus, taxon.Species, model.SpecimenPending); err != nil {
		return model.Specimen{}, err
	}
	specimen.Family, specimen.Genus, specimen.Species, specimen.Status = taxon.Family, taxon.Genus, taxon.Species, model.SpecimenPending
	return specimen, s.reports.Audit(ctx, "specimen", id, "classify", taxon.ScientificName())
}

func (s *SpecimenService) Accept(ctx context.Context, id int64) error {
	specimen, err := s.specimens.Get(ctx, id)
	if err != nil {
		return err
	}
	if !specimen.IsClassified() {
		return fmt.Errorf("%w: specimen has no classification", store.ErrState)
	}
	return s.specimens.UpdateClassification(ctx, id, specimen.Family, specimen.Genus, specimen.Species, model.SpecimenAccepted)
}

func (s *SpecimenService) Reject(ctx context.Context, id int64, note string) error {
	specimen, err := s.specimens.Get(ctx, id)
	if err != nil {
		return err
	}
	if specimen.Status == model.SpecimenAccepted {
		return fmt.Errorf("%w: accepted specimen cannot reject", store.ErrState)
	}
	if err := s.specimens.UpdateClassification(ctx, id, specimen.Family, specimen.Genus, specimen.Species, model.SpecimenRejected); err != nil {
		return err
	}
	return s.reports.Audit(ctx, "specimen", id, "reject", note)
}

func (s *SpecimenService) IndividualCount(ctx context.Context, batchID int64) (int, error) {
	return s.specimens.CountByBatch(ctx, batchID)
}
