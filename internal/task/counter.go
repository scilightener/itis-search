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

func NewDocumentCounterPipe(requiredDocumentCount int64, dirPath string, stopChan chan<- struct{}) pipe.Pipe[*Task] {
	return func(ctx context.Context, in <-chan *Task) <-chan *Task {
		out := make(chan *Task, cap(in))

		go func() {
			defer close(out)

			for t := range in {
				if err := ctx.Err(); err != nil {
					break
				}

				if t.Finished {
					out <- t
					continue
				}

				if counter.Load() < requiredDocumentCount {
					countDocuments(dirPath)
				}

				if counter.Load() >= requiredDocumentCount {
					once.Do(func() {
						stopChan <- struct{}{}
						close(stopChan)
					})

					continue
				}

				out <- t
			}
		}()

		return out
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
