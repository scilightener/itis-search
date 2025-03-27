package task

import (
	"context"
	"sync"
	"sync/atomic"

	"search/internal/pipe"
)

var (
	counter atomic.Int64
	once    sync.Once
)

func NewDocumentCounterFilter(requiredDocumentCount int64, stopFunc func()) pipe.Filter[*Task] {
	return func(ctx context.Context, t *Task) bool {
		if t.Finished {
			return true
		}

		if counter.Load() >= requiredDocumentCount {
			once.Do(stopFunc)
			return false
		}

		counter.Add(1)
		return true
	}
}
