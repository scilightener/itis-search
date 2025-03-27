package main

import (
	"context"
	"fmt"
	"search/internal/pkg"
	"time"

	"search/internal/index"
	"search/internal/pipe"
	"search/internal/task"
)

const (
	crawlerTimeout = time.Minute

	linksChanCapacity    = 100
	pipelineChanCapacity = 100

	numParallelFetchers = 20

	numDocumentWords = 1000
	numDocuments     = 100

	dataDirPath = "./data/5"
	indexPath   = "index5.json"
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
		pipe.NewPipe(task.NewSaveHandler(dataDirPath)),

		pipe.Synchronize(
			pipe.NewAsyncPipe(task.NewLogAsyncHandler()),
		),
		pipe.NewWaitPipe[*task.Task](),
	)

	time.Sleep(2 * time.Second)
	idx := index.NewIndex()
	err := idx.Load(indexPath)
	if err != nil {
		panic(err)
	}

	engine := index.NewSearchEngine(&idx.Data)
	query := "французский багет"
	query = pkg.NormalizeString(query)
	top := 5
	results := engine.Search(query, top)
	for _, result := range results {
		fmt.Printf("Документ %d: вес = %.4f\n", result.DocID, result.Score)
	}
}
