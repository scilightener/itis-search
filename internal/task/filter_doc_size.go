package task

import (
	"context"
	"strings"

	"search/internal/domain"
	"search/internal/pipe"
)

func NewDocumentSizeFilter(threshold int) pipe.Filter[*Task] {
	return func(_ context.Context, t *Task) bool {
		if t.Finished {
			return true
		}

		return !isSmallDocument(t.Document, threshold)
	}
}

func isSmallDocument(d domain.Document, threshold int) bool {
	text := string(d.Text)
	return len(strings.Fields(text)) < threshold
}
