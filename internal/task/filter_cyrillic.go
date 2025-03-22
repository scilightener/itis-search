package task

import (
	"context"
	"unicode"

	"search/internal/domain"
	"search/internal/pipe"
)

func NewFilterCyrillicPipe() pipe.Pipe[*Task] {
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

				if !isCyrillic(t.Document) {
					t = t.Failed("document is not cyrillic")
				}

				out <- t
			}
		}()

		return out
	}
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
