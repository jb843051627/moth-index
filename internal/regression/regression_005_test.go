package regression

import (
	"testing"

	"github.com/jb843051627/moth-index/internal/engine"
	"github.com/jb843051627/moth-index/internal/model"
)

func TestBug05_CacheResultsAreSnapshots(t *testing.T) {
	cache := engine.NewReadingCache()
	cache.Put(9, model.Reading{BatchID: 9, Temperature: 18, ObservedAt: "2026-08-24T01:00:00Z"})
	values := cache.Get(9)
	values[0].Temperature = 99
	all := cache.All()
	all[9][0].Humidity = 99
	stored := cache.Get(9)
	if stored[0].Temperature != 18 || stored[0].Humidity != 0 {
		t.Fatalf("cache was mutated through returned slice: %#v", stored[0])
	}
}
