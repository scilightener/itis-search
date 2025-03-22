package task

import (
	"context"
	"unicode"

	"search/internal/domain"
)

func CyrillicFilter(_ context.Context, t *Task) bool {
	if t.Finished {
		return true
	}

	return isCyrillic(t.Document)
}

func isCyrillic(d domain.Document) bool {
	text := []rune(string(d.Text))
	cyrillicCount := 0
	for _, r := range text {
		if unicode.Is(unicode.Cyrillic, r) {
			cyrillicCount++
		}
	}
	return float64(cyrillicCount)/float64(len(text)) >= 0.5
}
