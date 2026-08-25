package service

import (
	"context"
	"fmt"

	"github.com/jb843051627/moth-index/internal/model"
	"github.com/jb843051627/moth-index/internal/store"
)

type StationService struct {
	stations *store.StationStore
	traps    *store.TrapStore
	reports  *store.ReportStore
}

func NewStationService(stations *store.StationStore, traps *store.TrapStore, reports *store.ReportStore) *StationService {
	return &StationService{stations: stations, traps: traps, reports: reports}
}

func (s *StationService) Create(ctx context.Context, input model.Station) (model.Station, error) {
	input.Code = model.NormalizeStationCode(input.Code)
	input.Status = model.StationActive
	input.CreatedAt = serviceNow()
	if err := input.Validate(); err != nil {
		return model.Station{}, wrapValidation(err)
	}
	if _, err := s.stations.GetByCode(ctx, input.Code); err == nil {
		return model.Station{}, fmt.Errorf("%w: station code", store.ErrConflict)
	} else if !isNotFound(err) {
		return model.Station{}, err
	}
	return s.stations.Create(ctx, input)
}

func (s *StationService) Get(ctx context.Context, id int64) (model.Station, error) {
	return s.stations.Get(ctx, id)
}

func (s *StationService) List(ctx context.Context, page model.Page, includeRetired bool) ([]model.Station, error) {
	page = page.Normalize()
	return s.stations.List(ctx, page.Limit, page.Offset, includeRetired)
}

func (s *StationService) Retire(ctx context.Context, id int64) error {
	station, err := s.stations.Get(ctx, id)
	if err != nil {
		return err
	}
	if !station.CanRetire() {
		return fmt.Errorf("%w: station is not active", store.ErrState)
	}
	count, err := s.traps.CountByStation(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: station has traps", store.ErrConflict)
	}
	if err := s.stations.UpdateStatus(ctx, id, model.StationRetired); err != nil {
		return err
	}
	return s.reports.Audit(ctx, "station", id, "retire", "station retired")
}

func (s *StationService) Reactivate(ctx context.Context, id int64) error {
	station, err := s.stations.Get(ctx, id)
	if err != nil {
		return err
	}
	if station.Status != model.StationRetired {
		return fmt.Errorf("%w: station is not retired", store.ErrState)
	}
	if err := s.stations.UpdateStatus(ctx, id, model.StationActive); err != nil {
		return err
	}
	return s.reports.Audit(ctx, "station", id, "reactivate", "station reactivated")
}

func (s *StationService) TrapCount(ctx context.Context, id int64) (int, error) {
	if _, err := s.stations.Get(ctx, id); err != nil {
		return 0, err
	}
	return s.traps.CountByStation(ctx, id)
}

func (s *StationService) IsActive(ctx context.Context, id int64) (bool, error) {
	station, err := s.stations.Get(ctx, id)
	if err != nil {
		return false, err
	}
	return station.IsUsable(), nil
}
