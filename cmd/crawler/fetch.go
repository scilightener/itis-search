package main

import (
	"context"
	"time"

	"search/internal/pipe"
	"search/internal/task"
)

const (
	crawlerTimeout = time.Minute

	pipelineChanCapacity = 100
)

func fetch() {
	ctx, cancel := context.WithTimeout(context.Background(), crawlerTimeout)
	defer cancel()

	<-pipe.StartPipeline(ctx,
		task.NewTaskGeneratorFromStorage(pipelineChanCapacity, "./data/raw/", "./index.txt"),

		pipe.NewPipe(task.ProcessDocumentHandler),

		pipe.Synchronize(
			task.NewIndexerPipe(indexPath),
		),

		pipe.Synchronize(
			pipe.NewAsyncPipe(task.NewLogAsyncHandler()),
		),
		pipe.NewWaitPipe[*task.Task](),
	)
}
