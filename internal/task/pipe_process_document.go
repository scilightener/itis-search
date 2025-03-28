package task

import (
	"context"

	"search/internal/pkg"
)

func ProcessDocumentHandler(_ context.Context, t *Task) *Task {
	text := pkg.NormalizeString(string(t.Document.ProcessedText))
	t.Document.ProcessedText = []byte(text)

	return t
}
