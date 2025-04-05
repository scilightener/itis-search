package main

import (
	"context"
	"fmt"
	"time"

	"search/internal/index"
	"search/internal/pipe"
	"search/internal/task"
)

const (
	crawlerTimeout = time.Minute

	linksChanCapacity    = 100
	pipelineChanCapacity = 100

	dataDirPath = "./data/3"
	indexPath   = "index3.txt"
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
		task.NewTaskGeneratorFromStorage(pipelineChanCapacity, "./data/raw/", "./index.txt"),

		pipe.NewPipe(task.ProcessDocumentHandler),

		pipe.Synchronize(
			task.NewIndexerPipe(indexPath),
		),
		pipe.NewPipe(task.NewSaveHandler(dataDirPath)),

		pipe.Synchronize(
			pipe.NewAsyncPipe(task.NewLogAsyncHandler()),
		),
		pipe.NewWaitPipe[*task.Task](),
	)

	queries := []string{
		"авито & википедия | французский",
		"авито | википедия | французский",
		"авито & википедия & французский",
		"авито & !википедия | !французский",
		"авито | !википедия | !французский",
	}

	time.Sleep(2 * time.Second)
	idx := index.NewInverseIndex()
	err := idx.Load(indexPath)
	if err != nil {
		panic(err)
	}

	search := index.NewSearch(idx)

	for _, query := range queries {
		fmt.Printf("Search query: %s\n", query)
		fmt.Printf("Search results: %v\n", search.Search(query))
		fmt.Println()
	}
}
