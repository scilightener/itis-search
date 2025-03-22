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

	dataDirPath = "./data/1"
	indexPath   = "index.txt"
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
		pipe.NewParallelPipe(
			task.NewFetchPipe(), numParallelFetchers,
		),
		task.NewParsePipe(),
		task.NewFilterSmallDocumentsPipe(numDocumentWords),
		task.NewFilterCyrillicPipe(),
		pipe.NewPipeFromAsyncPipe(
			task.NewFeedLinksAsyncPipe(linksChan),
		),
		task.NewDocumentCounterPipe(numDocuments, dataDirPath, stopChan),
		task.NewIndexerPipe(indexPath),
		task.NewSavePipe(dataDirPath),
		pipe.NewPipeFromAsyncPipe(
			task.NewLogAsyncPipe(),
		),
		pipe.NewDiscardPipe[*task.Task](),
	)

	<-ctx.Done()
}
