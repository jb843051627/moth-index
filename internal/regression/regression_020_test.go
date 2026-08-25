package regression

import (
	"sync"
	"testing"
	"time"

	"github.com/jb843051627/moth-index/internal/engine"
	"github.com/jb843051627/moth-index/internal/model"
)

func TestBug20_CacheWritesAreSynchronized(t *testing.T) {
	cache := engine.NewReadingCache()
	var group sync.WaitGroup
	for worker := 0; worker < 6; worker++ {
		group.Add(1)
		go func(worker int) {
			defer group.Done()
			for i := 0; i < 100; i++ {
				cache.Put(int64(worker), model.Reading{BatchID: int64(worker), ObservedAt: time.Now().UTC().Format(time.RFC3339Nano), Temperature: float64(i)})
				cache.PruneBefore(time.Now().Add(-time.Hour))
				_ = cache.Get(int64(worker))
			}
		}(worker)
	}
	group.Wait()
}
