package task

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func FetchHandler(ctx context.Context, t *Task) *Task {
	const op = "task.fetch.FetchHandler"
	if t.Finished {
		return t
	}

	err := fetchTask(ctx, t)
	if err != nil {
		t = t.Fail(fmt.Sprintf("%s: %s", op, err.Error()))
	}

	return t
}

// fetchTask fetches a task and extracts its content.
func fetchTask(ctx context.Context, t *Task) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.Link, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
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
