package task

import (
	"context"

	"search/internal/pipe"
)

func NewFeedLinksAsyncHandler(linksChan chan<- string) pipe.AsyncHandler[*Task] {
	return func(ctx context.Context, t *Task) {
		maxLinks := min(len(t.Document.Links), cap(linksChan)-len(linksChan))

		for i := range maxLinks {
			linksChan <- t.Document.Links[i]
		}
	}
}
