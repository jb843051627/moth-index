package regression

import (
	"sync"
	"testing"

	"github.com/jb843051627/moth-index/internal/engine"
)

func TestBug01_MetricsConcurrentWrites(t *testing.T) {
	metrics := engine.NewMetrics()
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for j := 0; j < 100; j++ {
				metrics.Add("claimed", 1)
				metrics.Set("workers", int64(j))
			}
		}()
	}
	group.Wait()
	if metrics.Get("claimed") != 800 {
		t.Fatalf("claimed=%d, want 800", metrics.Get("claimed"))
	}
}
