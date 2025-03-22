package task

import (
	"context"
	"strings"

	"search/internal/domain"
	"search/internal/pipe"
)

func NewFilterSmallDocumentsPipe(threshold int) pipe.Pipe[*Task] {
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

				if isSmallDocument(t.Document, threshold) {
					t = t.Failed("document is too small")
				}

				out <- t
			}
		}()

		return out
	}
}

func isSmallDocument(d domain.Document, threshold int) bool {
	text := string(d.Text)
	return len(strings.Fields(text)) < threshold
}
