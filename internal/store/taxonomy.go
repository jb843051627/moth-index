package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jb843051627/moth-index/internal/model"
)

type TaxonomyStore struct{ db *DB }

func NewTaxonomyStore(db *DB) *TaxonomyStore { return &TaxonomyStore{db: db} }

func (s *TaxonomyStore) Upsert(ctx context.Context, taxon model.Taxon) (model.Taxon, error) {
	result, err := s.db.Exec(ctx, `INSERT INTO taxa(family,genus,species,common_name,authority) VALUES(?,?,?,?,?) ON CONFLICT(family,genus,species) DO UPDATE SET common_name=excluded.common_name,authority=excluded.authority`,
		taxon.Family, taxon.Genus, taxon.Species, taxon.Common, taxon.Authority)
	if err != nil {
		return model.Taxon{}, fmt.Errorf("upsert taxon: %w", err)
	}
	taxon.ID, err = txInsertID(result)
	if err != nil {
		existing, getErr := s.Find(ctx, taxon.Family, taxon.Genus, taxon.Species)
		if getErr != nil {
			return model.Taxon{}, err
		}
		return existing, nil
	}
	return taxon, nil
}

func (s *TaxonomyStore) Find(ctx context.Context, family, genus, species string) (model.Taxon, error) {
	var taxon model.Taxon
	err := s.db.QueryRow(ctx, `SELECT id,family,genus,species,common_name,authority FROM taxa WHERE family=? AND genus=? AND species=?`, family, genus, species).
		Scan(&taxon.ID, &taxon.Family, &taxon.Genus, &taxon.Species, &taxon.Common, &taxon.Authority)
	if err == sql.ErrNoRows {
		return model.Taxon{}, ErrNotFound
	}
	if err != nil {
		return model.Taxon{}, fmt.Errorf("find taxon: %w", err)
	}
	return taxon, nil
}

func (s *TaxonomyStore) Search(ctx context.Context, query string, limit int) ([]model.Taxon, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	pattern := "%" + query + "%"
	rows, err := s.db.Query(ctx, `SELECT id,family,genus,species,common_name,authority FROM taxa WHERE family LIKE ? OR genus LIKE ? OR species LIKE ? OR common_name LIKE ? ORDER BY family,genus,species LIMIT ?`, pattern, pattern, pattern, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("search taxa: %w", err)
	}
	values := make([]model.Taxon, 0)
	for rows.Next() {
		var taxon model.Taxon
		if err := rows.Scan(&taxon.ID, &taxon.Family, &taxon.Genus, &taxon.Species, &taxon.Common, &taxon.Authority); err != nil {
			return nil, closeRows(rows, fmt.Errorf("scan taxon: %w", err))
		}
		values = append(values, taxon)
	}
	return values, closeRows(rows, nil)
}

func (s *TaxonomyStore) Count(ctx context.Context) (int, error) {
	var count int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM taxa`).Scan(&count); err != nil {
		return 0, fmt.Errorf("count taxa: %w", err)
	}
	return count, nil
}
