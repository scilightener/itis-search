package task

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func FetchHandler(_ context.Context, t *Task) *Task {
	const op = "task.fetch.FetchHandler"
	if t.Finished {
		return t
	}

	err := fetchTask(t)
	if err != nil {
		t = t.Fail(fmt.Sprintf("%s: %s", op, err.Error()))
	}

	return t
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
