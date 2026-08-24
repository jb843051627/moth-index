package app

import (
	"context"
	"time"

	"github.com/jb843051627/moth-index/internal/model"
	"github.com/jb843051627/moth-index/internal/store"
)

func SmokeDatabase(path string) error {
	db, err := store.Open(path)
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	station := model.Station{Code: "SMOKE", Name: "Smoke Station", Timezone: "Asia/Shanghai", Status: model.StationActive, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	if _, err := store.NewStationStore(db).Create(ctx, station); err != nil {
		return err
	}
	return db.Ping(ctx)
}
