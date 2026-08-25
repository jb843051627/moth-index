package regression

import (
	"testing"

	"github.com/jb843051627/moth-index/internal/engine"
)

func TestBug06_QueueCloseIsIdempotent(t *testing.T) {
	queue := engine.NewQueue(nil, nil, 1)
	queue.Close()
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("second close panicked: %v", recovered)
		}
	}()
	queue.Close()
}
