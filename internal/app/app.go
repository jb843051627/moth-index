package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jb843051627/moth-index/internal/engine"
	"github.com/jb843051627/moth-index/internal/handler"
	"github.com/jb843051627/moth-index/internal/service"
	"github.com/jb843051627/moth-index/internal/store"
)

func Run(cfg Config) error {
	if !cfg.Valid() {
		return errors.New("invalid application configuration")
	}
	db, err := store.Open(cfg.DBPath)
	if err != nil {
		return err
	}
	defer db.Close()
	registry := service.NewRegistry(db)
	queue := engine.NewQueue(registry.Tasks, registry.Reviews, cfg.Workers)
	defer queue.Close()
	server := &http.Server{Addr: cfg.Address, Handler: handler.NewRouter(registry, queue), ReadHeaderTimeout: 5 * time.Second}
	return serveUntilSignal(server)
}

func serveUntilSignal(server *http.Server) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	errorsCh := make(chan error, 1)
	go func() { errorsCh <- server.ListenAndServe() }()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}
		return nil
	case err := <-errorsCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve: %w", err)
	}
}
