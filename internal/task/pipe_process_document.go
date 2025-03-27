package task

import (
	"context"

	"search/internal/pkg"
)

func ProcessDocumentHandler(_ context.Context, t *Task) *Task {
	text := pkg.NormalizeString(string(t.Document.Text))
	t.Document.Text = []byte(text)

	return t
}
