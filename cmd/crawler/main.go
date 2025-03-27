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

	numParallelFetchers = 20

	numDocumentWords = 1000
	numDocuments     = 100

	dataDirPath = "./data/4"
	indexPath   = "index4.json"
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
	idx := index.NewInverseIndex()
	err := idx.Load(indexPath)
	if err != nil {
		panic(err)
	}

	if err := idx.SaveTF("tf.csv"); err != nil {
		panic(err)
	}
	if err := idx.SaveIDF("idf.csv"); err != nil {
		panic(err)
	}
	if err := idx.SaveTFIDF("tfidf.csv"); err != nil {
		panic(err)
	}

	fmt.Println("Saved tables: tf.csv, idf.csv, tfidf.csv")

	tf := idx.CalculateTF()
	idf := idx.CalculateIDF()
	tfidf := idx.CalculateTFIDF()

	idfData := make(map[string]map[int64]float64)
	for word, values := range tf {
		idfData[word] = make(map[int64]float64)
		for docID := range values {
			idfData[word][docID] = idf[word]
		}
	}

	maxRows := 10
	index.PrintTable("TF (Term Frequency)", tf, maxRows)
	fmt.Println()
	index.PrintTable("IDF (Inverse Document Frequency)", idfData, maxRows)
	fmt.Println()
	index.PrintTable("TF-IDF", tfidf, maxRows)
	fmt.Println()
}
