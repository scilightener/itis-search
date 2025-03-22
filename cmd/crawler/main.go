package main

import (
	"context"
	"time"

	"search/internal/pipe"
	"search/internal/task"
)

const (
	crawlerTimeout = time.Minute

	linksChanCapacity    = 1000
	pipelineChanCapacity = 1000

	numParallelFetchers = 20

	numDocumentWords = 1000
	numDocuments     = 200

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

	stopChan := make(chan struct{})

	_ = pipe.StartPipeline(ctx,
		task.NewTaskGenerator(pipelineChanCapacity, linksChan, stopChan),

		pipe.Parallelize(
			pipe.NewPipe(task.FetchHandler), numParallelFetchers,
		),
		pipe.NewPipe(task.ParseHandler),

		pipe.Satisfies(task.NewBigDocumentFilter(numDocumentWords)),
		pipe.Satisfies(task.CyrillicFilter),
		pipe.Satisfies(task.NewDocumentCounterFilter(numDocuments, dataDirPath, stopChan)),

		pipe.NewPipe(task.ProcessDocumentHandler),

		pipe.Synchronize(
			pipe.NewAsyncPipe(task.NewFeedLinksAsyncHandler(linksChan)),
		),

		pipe.NewPipe(task.NewIndexerHandler(indexPath)),
		pipe.NewPipe(task.NewSaveHandler(dataDirPath)),

		pipe.Synchronize(
			pipe.NewAsyncPipe(task.NewLogAsyncHandler()),
		),
		pipe.NewDiscardPipe[*task.Task](),
	)

	<-ctx.Done()
}
