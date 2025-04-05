package main

import (
	"context"
	"time"

	"search/internal/pipe"
	"search/internal/task"
)

const (
	crawlerTimeout = time.Minute

	linksChanCapacity    = 100
	pipelineChanCapacity = 100

	dataDirPath = "./data/2"
	indexPath   = "index2.txt"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), crawlerTimeout)
	defer cancel()

	links := []string{
		"https://ru.wikipedia.org/wiki/Авито",
		"https://ru.wikipedia.org/wiki/Французский_язык",
	}

	linksChan := make(chan string, linksChanCapacity)
	for _, link := range links {
		linksChan <- link
	}

	_ = pipe.StartPipeline(ctx,
		task.NewTaskGeneratorFromStorage(pipelineChanCapacity, "./data/raw/", "./index.txt"),

		pipe.NewPipe(task.ProcessDocumentHandler),

		pipe.NewPipe(task.NewIndexerHandler(indexPath)),
		pipe.NewPipe(task.NewSaveHandler(dataDirPath)),

		pipe.Synchronize(
			pipe.NewAsyncPipe(task.NewLogAsyncHandler()),
		),
		pipe.NewDiscardPipe[*task.Task](),
	)

	<-ctx.Done()
}
