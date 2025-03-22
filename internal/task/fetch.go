package task

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"search/internal/pipe"
)

// NewFetchPipe returns a new pipe that fetches tasks.
func NewFetchPipe() pipe.Pipe[*Task] {
	const op = "task.fetch.NewFetchPipe"

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

				err := fetchTask(t)
				if err != nil {
					t = t.Failed(fmt.Sprintf("%s: %s", op, err.Error()))
				}

				out <- t
			}
		}()

		return out
	}
}

// fetchTask fetches a task and extracts its content.
func fetchTask(t *Task) error {
	resp, err := http.Get(t.Link)
	if err != nil {
		return fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch URL: %s", resp.Status)
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	t.Document.Text = respBytes

	return nil
}
