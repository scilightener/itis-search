package task

import (
	"context"
	"fmt"
	"os"
	"path"
	"strconv"

	"search/internal/pipe"
)

func NewSavePipe(dirPath string) pipe.Pipe[*Task] {
	const op = "task.save.NewSavePipe"

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

				err := saveDocument(t, dirPath)
				if err != nil {
					t = t.Failed(fmt.Sprintf("%s: %s", op, err.Error()))
				}

				out <- t
			}
		}()

		return out
	}
}

func saveDocument(t *Task, dirPath string) error {
	err := ensureDirExists(dirPath)
	if err != nil {
		return err
	}

	fileName := strconv.FormatInt(t.ID, 10) + ".txt"
	filePath := path.Join(dirPath, fileName)
	if err := os.WriteFile(filePath, t.Document.Text, 0644); err != nil {
		return fmt.Errorf("failed to save document: %w", err)
	}

	return nil
}

func ensureDirExists(dirPath string) error {
	err := os.MkdirAll(dirPath, 0774)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return nil
}
