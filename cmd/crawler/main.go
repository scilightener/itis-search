package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/fatih/color"

	"search/internal/index"
	"search/internal/pipe"
	"search/internal/pkg"
	"search/internal/task"
)

const (
	crawlerTimeout = time.Minute

	linksChanCapacity    = 100
	pipelineChanCapacity = 100

	numParallelFetchers = 20

	numDocumentWords = 1000
	numDocuments     = 200

	processedDocumentsDirPath = "./data/5/processed"
	rawDocumentsDirPath       = "./data/5/raw"
	indexPath                 = "index5.json"
)

func main() {
	//fetch()
	//time.Sleep(2 * time.Second)

	idx := index.NewIndex()
	err := idx.Load(indexPath)
	if err != nil {
		color.Red("Ошибка загрузки индекса: %v\n", err)
		return
	}

	engine := index.NewSearchEngine(&idx.Data)
	top := 5

	color.Cyan("Поисковая система запущена. Введите запрос (или 'exit' для выхода):")
	color.Yellow("Доступные команды: :top N, :clear")

	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "\033[32mПоиск>\033[0m ",
		HistoryFile:  "/tmp/search_history.txt",
		HistoryLimit: 100,
	})
	if err != nil {
		panic(err)
	}
	defer func(rl *readline.Instance) {
		err := rl.Close()
		if err != nil {
			color.Red("Ошибка закрытия readline: %v", err)
		}
	}(rl)

	completer := readline.NewPrefixCompleter(
		readline.PcItem(":clear"),
		readline.PcItem(":top"),
		readline.PcItem("exit"),
	)
	rl.Config.AutoComplete = completer

	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}
		query := strings.TrimSpace(line)

		switch {
		case strings.ToLower(query) == "exit":
			color.Green("Завершение работы.")
			return

		case strings.ToLower(query) == ":clear":
			rl.ResetHistory()
			if err := rl.SaveHistory(query); err != nil {
				color.Red("Ошибка сохранения истории: %v", err)
			}
			color.Green("История очищена")
			continue

		case strings.HasPrefix(query, ":top "):
			_, err := fmt.Sscanf(query, ":top %d", &top)
			if err != nil {
				color.Red("Неверный формат. Используйте: :top N")
			} else {
				color.Green("Установлено количество результатов: %d", top)
			}
			continue
		}

		if query == "" {
			continue
		}

		if err := rl.SaveHistory(query); err != nil {
			color.Red("Ошибка сохранения истории: %v", err)
		}

		results := engine.Search(query, top)
		printResults(query, results)
	}
}

func printResults(query string, results []index.SearchResult) {
	normalizedQuery := pkg.NormalizeString(query)
	queryWords := strings.Fields(normalizedQuery)

	yellow := color.New(color.FgYellow).SprintFunc()
	fmt.Printf("\nРезультатов по запросу '%s': %d\n", query, len(results))

	for i, result := range results {
		snippet := result.Snippet
		originalWords := strings.Fields(snippet)

		highlightPositions := make(map[int]bool, len(originalWords))

		for wordPos, originalWord := range originalWords {
			normalizedWord := pkg.NormalizeString(originalWord)

			for _, qWord := range queryWords {
				if normalizedWord == qWord ||
					(len(qWord) > 3 && strings.Contains(normalizedWord, qWord)) ||
					(len(normalizedWord) > 3 && strings.Contains(qWord, normalizedWord)) {
					highlightPositions[wordPos] = true
					break
				}
			}
		}

		var builder strings.Builder
		for wordPos, word := range originalWords {
			if highlightPositions[wordPos] {
				builder.WriteString(yellow(word))
			} else {
				builder.WriteString(word)
			}

			if wordPos < len(originalWords)-1 {
				builder.WriteString(" ")
			}
		}

		color.Blue("%d. Документ %d (релевантность: %.2f)", i+1, result.DocID, result.Score)
		fmt.Printf("   Сниппет: %s\n", builder.String())
		color.White("   " + strings.Repeat("─", 60))
	}
}

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
