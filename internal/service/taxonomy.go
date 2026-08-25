package service

import (
	"context"
	"strings"

	"github.com/jb843051627/moth-index/internal/model"
	"github.com/jb843051627/moth-index/internal/store"
)

type TaxonomyService struct {
	taxa      *store.TaxonomyStore
	specimens *store.SpecimenStore
}

func NewTaxonomyService(taxa *store.TaxonomyStore, specimens *store.SpecimenStore) *TaxonomyService {
	return &TaxonomyService{taxa: taxa, specimens: specimens}
}

func (s *TaxonomyService) Register(ctx context.Context, taxon model.Taxon) (model.Taxon, error) {
	taxon.Family = strings.TrimSpace(taxon.Family)
	taxon.Genus = strings.TrimSpace(taxon.Genus)
	taxon.Species = strings.TrimSpace(taxon.Species)
	if err := taxon.Validate(); err != nil {
		return model.Taxon{}, wrapValidation(err)
	}
	return s.taxa.Upsert(ctx, taxon)
}

func (s *TaxonomyService) Find(ctx context.Context, family, genus, species string) (model.Taxon, error) {
	return s.taxa.Find(ctx, strings.TrimSpace(family), strings.TrimSpace(genus), strings.TrimSpace(species))
}

func (s *TaxonomyService) Search(ctx context.Context, query string, limit int) ([]model.Taxon, error) {
	return s.taxa.Search(ctx, query, limit)
}

func (s *TaxonomyService) Suggest(ctx context.Context, specimenID int64) ([]model.Taxon, error) {
	specimen, err := s.specimens.Get(ctx, specimenID)
	if err != nil {
		return nil, err
	}
	query := specimen.Genus
	if query == "" {
		query = specimen.Family
	}
	return s.taxa.Search(ctx, query, 10)
}

func (s *TaxonomyService) Exists(ctx context.Context, family, genus, species string) (bool, error) {
	_, err := s.taxa.Find(ctx, family, genus, species)
	if err == store.ErrNotFound {
		return false, nil
	}
	return err == nil, err
}

func (s *TaxonomyService) Count(ctx context.Context) (int, error) { return s.taxa.Count(ctx) }
