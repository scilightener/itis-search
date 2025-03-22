package main

import (
	"context"
	"time"

	"search/internal/pipe"
	"search/internal/task"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	links := []string{
		"https://ru.wikipedia.org/wiki/Авито",
		"https://ru.wikipedia.org/wiki/Французский_язык",
	}

	linksChan := make(chan string, 100)
	for _, link := range links {
		linksChan <- link
	}

	stopChan := make(chan struct{})

	_ = pipe.StartPipeline(ctx,
		task.NewTaskGenerator(100, linksChan, stopChan),
		pipe.NewParallelPipe(
			task.NewFetchPipe(), 20,
		),
		task.NewParsePipe(),
		task.NewFilterSmallDocumentsPipe(1000),
		task.NewDocumentCounterPipe(200, "./data/1", stopChan),
		task.NewIndexerPipe("index.txt"),
		pipe.NewPipeFromAsyncPipe(
			task.NewFeedLinksAsyncPipe(linksChan),
		),
		task.NewSavePipe("./data/1"),
		pipe.NewPipeFromAsyncPipe(
			task.NewLogAsyncPipe(),
		),
		pipe.NewDiscardPipe[*task.Task](),
	)

	<-ctx.Done()
}
