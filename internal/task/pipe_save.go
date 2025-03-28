package task

import (
	"context"
	"fmt"
	"os"
	"path"
	"strconv"

	"search/internal/pipe"
)

func NewSaveHandler(processedDirPath, rawDirPath string) pipe.Handler[*Task] {
	const op = "task.save.NewSaveHandler"

	err := emptyDir(processedDirPath)
	if err != nil {
		panic(err)
	}
	err = ensureDirExists(processedDirPath)
	if err != nil {
		panic(err)
	}

	err = emptyDir(rawDirPath)
	if err != nil {
		panic(err)
	}
	err = ensureDirExists(rawDirPath)
	if err != nil {
		panic(err)
	}

	return func(ctx context.Context, t *Task) *Task {
		if t.Finished {
			return t
		}

		err := saveDocument(t.Document.ID, t.Document.ProcessedText, processedDirPath)
		if err != nil {
			t = t.Fail(fmt.Sprintf("%s: %s", op, err.Error()))
			return t
		}

		err = saveDocument(t.Document.ID, t.Document.RawText, rawDirPath)
		if err != nil {
			t = t.Fail(fmt.Sprintf("%s: %s", op, err.Error()))
			return t
		}

		return t
	}
}

func saveDocument(id int64, content []byte, dirPath string) error {
	fileName := strconv.FormatInt(id, 10) + ".txt"
	filePath := path.Join(dirPath, fileName)
	if err := os.WriteFile(filePath, content, 0644); err != nil {
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
