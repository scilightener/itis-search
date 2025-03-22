package task

import (
	"context"

	"search/internal/pipe"
)

func NewFeedLinksAsyncPipe(linksChan chan<- string) pipe.AsyncPipe[*Task] {
	return func(ctx context.Context, in <-chan *Task) {
		go func() {
			for t := range in {
				if err := ctx.Err(); err != nil {
					break
				}

				for _, link := range t.Document.Links {
					linksChan <- link
				}
			}
		}()
	}
}
