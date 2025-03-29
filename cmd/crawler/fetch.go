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

	numParallelFetchers = 50

	numDocumentWords = 1000
	numDocuments     = 200
)

func fetch() {
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

	<-pipe.StartPipeline(ctx,
		task.NewTaskGenerator(pipelineChanCapacity, linksChan),

		pipe.Parallelize(
			pipe.NewPipe(task.FetchHandler), numParallelFetchers,
		),
		pipe.NewPipe(task.ParseHandler),

		pipe.Satisfies(task.NewDocumentSizeFilter(numDocumentWords)),
		pipe.Satisfies(task.CyrillicFilter),
		pipe.Satisfies(task.NewDocumentCounterFilter(numDocuments, cancel)),
		pipe.Satisfies(task.NotFinishedFilter),

		pipe.NewPipe(task.ProcessDocumentHandler),

		pipe.Synchronize(
			pipe.NewAsyncPipe(task.NewFeedLinksAsyncHandler(linksChan)),
		),

		pipe.Synchronize(
			task.NewIndexerPipe(indexPath),
		),
		pipe.NewPipe(task.NewSaveHandler(processedDocumentsDirPath, rawDocumentsDirPath)),

		pipe.Synchronize(
			pipe.NewAsyncPipe(task.NewLogAsyncHandler()),
		),
		pipe.NewWaitPipe[*task.Task](),
	)
}
