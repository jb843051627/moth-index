package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jb843051627/moth-index/internal/model"
)

type SpecimenStore struct{ db *DB }

func NewSpecimenStore(db *DB) *SpecimenStore { return &SpecimenStore{db: db} }

func (s *SpecimenStore) Create(ctx context.Context, specimen model.Specimen) (model.Specimen, error) {
	result, err := s.db.Exec(ctx, `INSERT INTO specimens(batch_id,tag,family,genus,species,sex,count,status,notes,created_at) VALUES(?,?,?,?,?,?,?,?,?,?)`,
		specimen.BatchID, specimen.Tag, specimen.Family, specimen.Genus, specimen.Species, specimen.Sex, specimen.Count, specimen.Status, specimen.Notes, specimen.CreatedAt)
	if err != nil {
		return model.Specimen{}, fmt.Errorf("insert specimen: %w", err)
	}
	specimen.ID, err = txInsertID(result)
	if err != nil {
		return model.Specimen{}, err
	}
	return specimen, nil
}

func (s *SpecimenStore) Get(ctx context.Context, id int64) (model.Specimen, error) {
	var specimen model.Specimen
	err := s.db.QueryRow(ctx, `SELECT id,batch_id,tag,family,genus,species,sex,count,status,notes,created_at FROM specimens WHERE id=?`, id).
		Scan(&specimen.ID, &specimen.BatchID, &specimen.Tag, &specimen.Family, &specimen.Genus, &specimen.Species, &specimen.Sex, &specimen.Count, &specimen.Status, &specimen.Notes, &specimen.CreatedAt)
	if err == sql.ErrNoRows {
		return model.Specimen{}, ErrNotFound
	}
	if err != nil {
		return model.Specimen{}, fmt.Errorf("get specimen: %w", err)
	}
	return specimen, nil
}

func (s *SpecimenStore) GetByTag(ctx context.Context, tag string) (model.Specimen, error) {
	var specimen model.Specimen
	err := s.db.QueryRow(ctx, `SELECT id,batch_id,tag,family,genus,species,sex,count,status,notes,created_at FROM specimens WHERE tag=?`, model.NormalizeTag(tag)).
		Scan(&specimen.ID, &specimen.BatchID, &specimen.Tag, &specimen.Family, &specimen.Genus, &specimen.Species, &specimen.Sex, &specimen.Count, &specimen.Status, &specimen.Notes, &specimen.CreatedAt)
	if err == sql.ErrNoRows {
		return model.Specimen{}, ErrNotFound
	}
	if err != nil {
		return model.Specimen{}, fmt.Errorf("get specimen by tag: %w", err)
	}
	return specimen, nil
}

func (s *SpecimenStore) ListByBatch(ctx context.Context, batchID int64, limit, offset int) ([]model.Specimen, error) {
	limit, offset = pageValues(limit, offset)
	rows, err := s.db.Query(ctx, `SELECT id,batch_id,tag,family,genus,species,sex,count,status,notes,created_at FROM specimens WHERE batch_id=? ORDER BY id LIMIT ? OFFSET ?`, batchID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list specimens: %w", err)
	}
	values := make([]model.Specimen, 0)
	for rows.Next() {
		var specimen model.Specimen
		if err := rows.Scan(&specimen.ID, &specimen.BatchID, &specimen.Tag, &specimen.Family, &specimen.Genus, &specimen.Species, &specimen.Sex, &specimen.Count, &specimen.Status, &specimen.Notes, &specimen.CreatedAt); err != nil {
			return nil, closeRows(rows, fmt.Errorf("scan specimen: %w", err))
		}
		values = append(values, specimen)
	}
	return values, closeRows(rows, nil)
}

func (s *SpecimenStore) UpdateClassification(ctx context.Context, id int64, family, genus, species, status string) error {
	result, err := s.db.Exec(ctx, `UPDATE specimens SET family=?,genus=?,species=?,status=? WHERE id=?`, family, genus, species, status, id)
	if err != nil {
		return fmt.Errorf("update specimen classification: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read specimen update: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SpecimenStore) CountByBatch(ctx context.Context, batchID int64) (int, error) {
	var count int
	if err := s.db.QueryRow(ctx, `SELECT COALESCE(SUM(count),0) FROM specimens WHERE batch_id=?`, batchID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count specimen individuals: %w", err)
	}
	return count, nil
}
