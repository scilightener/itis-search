package task

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"search/internal/pipe"
)

func NewTaskGeneratorFromStorage(chanCapacity int, storagePath, indexPath string) pipe.Source[*Task] {
	storageDir, err := os.ReadDir(storagePath)
	if err != nil {
		panic(err)
	}

	indexFile, err := os.ReadFile(indexPath)
	if err != nil {
		panic(err)
	}
	id2link := make(map[int64]string)
	for _, line := range strings.Split(string(indexFile), "\n") {
		parts := strings.Split(line, " ")
		id, err := strconv.Atoi(parts[0])
		if err != nil {
			fmt.Println("ой-ёй", err)
			continue
		}
		link := parts[1]
		id2link[int64(id)] = link
	}

	return func(ctx context.Context) <-chan *Task {
		out := make(chan *Task, chanCapacity)

		go func() {
			defer close(out)

			for _, file := range storageDir {
				idStr := strings.TrimSuffix(file.Name(), ".txt")
				id, err := strconv.Atoi(idStr)
				if err != nil {
					fmt.Println("ой-ёй 2", err)
					continue
				}

				task := NewTask(int64(id), id2link[int64(id)])
				doc, err := os.ReadFile(storagePath + file.Name())
				if err != nil {
					fmt.Println("ой-ёй 3", err)
				}
				task.Document.RawText = doc
				task.Document.ProcessedText = doc
				task.Document.ID = int64(id)

				out <- task
			}
		}()

		return out
	}
}
