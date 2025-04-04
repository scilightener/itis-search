package task

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"search/internal/pipe"
)

func NewIndexerHandler(indexFileName string) pipe.Handler[*Task] {
	const op = "task.index.NewIndexerHandler"

	err := deleteIndexIfExists(indexFileName)
	if err != nil {
		panic(err)
	}

	return func(ctx context.Context, t *Task) *Task {
		if t.Finished {
			return t
		}

		err := saveToIndex(t, indexFileName)
		if err != nil {
			t = t.Fail(fmt.Sprintf("%s: %s", op, err.Error()))
		}

		return t
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

	link := unescapeLink(t.Document.URI)
	_, err = file.WriteString(fmt.Sprintf("%d %s\n", t.ID, link))
	if err != nil {
		return err
	}

	return nil
}

func unescapeLink(link string) string {
	unescaped, err := url.QueryUnescape(link)
	if err != nil {
		return link
	}

	return unescaped
}
