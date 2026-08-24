package service

import (
	"context"
	"fmt"

	"github.com/jb843051627/moth-index/internal/model"
	"github.com/jb843051627/moth-index/internal/store"
)

type TrapService struct {
	traps    *store.TrapStore
	stations *store.StationStore
	batches  *store.BatchStore
}

func NewTrapService(traps *store.TrapStore, stations *store.StationStore, batches *store.BatchStore) *TrapService {
	return &TrapService{traps: traps, stations: stations, batches: batches}
}

func (s *TrapService) Register(ctx context.Context, input model.Trap) (model.Trap, error) {
	input.Status = model.TrapReady
	input.Installed = serviceNow()
	if err := input.Validate(); err != nil {
		return model.Trap{}, wrapValidation(err)
	}
	station, err := s.stations.Get(ctx, input.StationID)
	if err != nil {
		return model.Trap{}, err
	}
	if !station.IsUsable() {
		return model.Trap{}, fmt.Errorf("%w: station is retired", store.ErrState)
	}
	return s.traps.Create(ctx, input)
}

func (s *TrapService) Get(ctx context.Context, id int64) (model.Trap, error) {
	return s.traps.Get(ctx, id)
}

func (s *TrapService) List(ctx context.Context, stationID int64, page model.Page) ([]model.Trap, error) {
	if _, err := s.stations.Get(ctx, stationID); err != nil {
		return nil, err
	}
	page = page.Normalize()
	return s.traps.ListByStation(ctx, stationID, page.Limit, page.Offset)
}

func (s *TrapService) Deploy(ctx context.Context, id int64) error {
	trap, err := s.traps.Get(ctx, id)
	if err != nil {
		return err
	}
	if trap.Status != model.TrapReady {
		return fmt.Errorf("%w: trap is not ready", store.ErrState)
	}
	return s.traps.SetStatus(ctx, id, model.TrapDeployed)
}

func (s *TrapService) Release(ctx context.Context, id int64) error {
	trap, err := s.traps.Get(ctx, id)
	if err != nil {
		return err
	}
	if trap.Status != model.TrapDeployed {
		return fmt.Errorf("%w: trap is not deployed", store.ErrState)
	}
	open, err := s.batches.CountOpenByTrap(ctx, id)
	if err != nil {
		return err
	}
	if open > 0 {
		return fmt.Errorf("%w: trap has open batch", store.ErrConflict)
	}
	return s.traps.SetStatus(ctx, id, model.TrapReady)
}

func (s *TrapService) Break(ctx context.Context, id int64) error {
	if _, err := s.traps.Get(ctx, id); err != nil {
		return err
	}
	return s.traps.SetStatus(ctx, id, model.TrapFault)
}

func (s *TrapService) CanUse(ctx context.Context, id int64) (bool, error) {
	trap, err := s.traps.Get(ctx, id)
	if err != nil {
		return false, err
	}
	station, err := s.stations.Get(ctx, trap.StationID)
	if err != nil {
		return false, err
	}
	return trap.IsDeployable() && station.IsUsable(), nil
}
