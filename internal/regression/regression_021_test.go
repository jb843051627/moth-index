package regression

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jb843051627/moth-index/internal/engine"
)

func TestBug21_RetryReturnsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started := time.Now()
	err := engine.Retry(ctx, 3, 300*time.Millisecond, func(context.Context) error {
		cancel()
		return errors.New("temporary field error")
	})
	if !errors.Is(err, context.Canceled) { t.Fatalf("retry error=%v", err) }
	if elapsed := time.Since(started); elapsed > 150*time.Millisecond { t.Fatalf("retry ignored cancellation for %s", elapsed) }
}
