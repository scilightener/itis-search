package task

import (
	"context"
	"fmt"
	"os"
	"path"
	"strconv"

	"search/internal/pipe"
)

func NewSaveHandler(dirPath string) pipe.Handler[*Task] {
	const op = "task.save.NewSaveHandler"

	err := emptyDir(dirPath)
	if err != nil {
		panic(err)
	}
	err = ensureDirExists(dirPath)
	if err != nil {
		panic(err)
	}

	return func(ctx context.Context, t *Task) *Task {
		if t.Finished {
			return t
		}

		err := saveDocument(t, dirPath)
		if err != nil {
			t = t.Fail(fmt.Sprintf("%s: %s", op, err.Error()))
		}

		return t
	}
}

func saveDocument(t *Task, dirPath string) error {
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

func emptyDir(dirPath string) error {
	err := os.RemoveAll(dirPath)
	if err != nil {
		return fmt.Errorf("failed to remove directory: %w", err)
	}

	return nil
}
