package task

import (
	"context"
	"log"
	"os"
	"regexp"
	"sync"
	"sync/atomic"

	"search/internal/pipe"
)

var (
	counter atomic.Int64
	once    sync.Once
)

func NewDocumentCounterFilter(requiredDocumentCount int64, dirPath string, stopChan chan<- struct{}) pipe.Filter[*Task] {
	return func(ctx context.Context, t *Task) bool {
		if t.Finished {
			return true
		}

		if counter.Load() < requiredDocumentCount {
			countDocuments(dirPath)
		}

		if counter.Load() >= requiredDocumentCount {
			once.Do(func() {
				stopChan <- struct{}{}
				close(stopChan)
			})

			return false
		}

		return true
	}
}

func countDocuments(dirPath string) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		log.Printf("error: %s\n", err.Error())
		return
	}

	re := regexp.MustCompile(`^\d+\.txt$`)

	count := int64(0)
	for _, file := range files {
		if re.MatchString(file.Name()) {
			count++
		}
	}

	counter.Store(count)
}
