package task

import (
	"context"
	"fmt"
	"os"

	"search/internal/pipe"
)

func NewIndexerPipe(indexFileName string) pipe.Pipe[*Task] {
	const op = "task.index.NewIndexerPipe"

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

				err := saveToIndex(t, indexFileName)
				if err != nil {
					t = t.Failed(fmt.Sprintf("%s: %s", op, err.Error()))
				}

				out <- t
			}
		}()

		return out
	}
}

func saveToIndex(t *Task, fileName string) error {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0774)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	_, err = file.WriteString(fmt.Sprintf("%d %s\n", t.ID, t.Document.URI))
	if err != nil {
		return err
	}

	return nil
}
