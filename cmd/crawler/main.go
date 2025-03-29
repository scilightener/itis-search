package main

import (
	"fmt"
	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"os"
	"strings"

	"search/internal/index"
	"search/internal/pkg"
)

const (
	processedDocumentsDirPath = "./data/processed"
	rawDocumentsDirPath       = "./data/raw"
	indexPath                 = "index5.json"

	historyFile  = "/tmp/search_history.txt"
	historyLimit = 100
)

type SearchApp struct {
	engine     *index.SearchEngine
	top        int
	windowSize int
}

func main() {
	app := initializeApp()
	app.runSearchLoop()
}

func initializeApp() *SearchApp {
	idx := index.NewIndex()
	if err := idx.Load(indexPath); err != nil {
		color.Red("Ошибка загрузки индекса: %v\n", err)
		os.Exit(1)
	}

	return &SearchApp{
		engine:     index.NewSearchEngine(&idx.Data),
		top:        5,
		windowSize: 20,
	}
}

func (app *SearchApp) runSearchLoop() {
	color.Cyan("Поисковая система запущена. Введите запрос (или 'exit' для выхода):")
	color.Yellow("Доступные команды: :top N, :window N, :clear")

	rl := app.setupReadline()
	defer func(rl *readline.Instance) {
		err := rl.Close()
		if err != nil {
			color.Red("Ошибка сохранения истории: %v", err)
		}
	}(rl)

	for {
		query, shouldExit := app.readInput(rl)
		if shouldExit {
			color.Green("Завершение работы.")
			return
		}

		shouldExit = app.processQuery(query, rl)
		if shouldExit {
			color.Green("Завершение работы.")
			return
		}
	}
}

func (app *SearchApp) setupReadline() *readline.Instance {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "\033[32mПоиск>\033[0m ",
		HistoryFile:  historyFile,
		HistoryLimit: historyLimit,
	})
	if err != nil {
		panic(err)
	}

	rl.Config.AutoComplete = readline.NewPrefixCompleter(
		readline.PcItem(":clear"),
		readline.PcItem(":top"),
		readline.PcItem(":window"),
		readline.PcItem("exit"),
	)

	return rl
}

func (app *SearchApp) readInput(rl *readline.Instance) (string, bool) {
	line, err := rl.Readline()
	if err != nil {
		return "", true
	}
	return strings.TrimSpace(line), false
}

func (app *SearchApp) processQuery(query string, rl *readline.Instance) (shouldReturn bool) {
	switch {
	case strings.ToLower(query) == "exit":
		return true

	case strings.ToLower(query) == ":clear":
		app.clearHistory(rl)
		return false

	case strings.HasPrefix(query, ":top "):
		app.setTopResults(query)
		return false

	case strings.HasPrefix(query, ":window "):
		app.setContextSize(query)
		return false

	case query == "":
		return false
	}

	if err := rl.SaveHistory(query); err != nil {
		color.Red("Ошибка сохранения истории: %v", err)
	}

	app.showSearchResults(query)
	return false
}

func (app *SearchApp) clearHistory(rl *readline.Instance) {
	rl.ResetHistory()
	if err := rl.SaveHistory(""); err != nil {
		color.Red("Ошибка сохранения истории: %v", err)
	}
	color.Green("История очищена")
}

func (app *SearchApp) setTopResults(query string) {
	_, err := fmt.Sscanf(query, ":top %d", &app.top)
	if err != nil {
		color.Red("Неверный формат. Используйте: :top N")
	} else {
		color.Green("Установлено количество результатов: %d", app.top)
	}
}

func (app *SearchApp) setContextSize(query string) {
	_, err := fmt.Sscanf(query, ":window %d", &app.windowSize)
	if err != nil {
		color.Red("Неверный формат. Используйте: :window N")
	} else {
		color.Green("Установлено количество результатов: %d", app.windowSize)
	}
}

func (app *SearchApp) showSearchResults(query string) {
	results := app.engine.Search(query, app.top, app.windowSize)
	printResults(query, results)
}

func printResults(query string, results []index.SearchResult) {
	normalizedQuery := pkg.NormalizeString(query)
	queryWords := strings.Fields(normalizedQuery)
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Printf("\nРезультатов по запросу '%s': %d\n", query, len(results))

	for i, result := range results {
		printSingleResult(i+1, result, queryWords, yellow)
	}
}

func printSingleResult(position int, result index.SearchResult, queryWords []string, highlightFunc func(a ...interface{}) string) {
	snippet := result.Snippet
	originalWords := strings.Fields(snippet)
	highlightPositions := findHighlightPositions(originalWords, queryWords)

	var builder strings.Builder
	for wordPos, word := range originalWords {
		if highlightPositions[wordPos] {
			builder.WriteString(highlightFunc(word))
		} else {
			builder.WriteString(word)
		}

		if wordPos < len(originalWords)-1 {
			builder.WriteString(" ")
		}
	}

	color.Blue("%d. Документ %d (релевантность: %.2f)", position, result.DocID, result.Score)
	fmt.Printf("   Сниппет: %s\n", builder.String())
	color.White("   " + strings.Repeat("─", 60))
}

func findHighlightPositions(words []string, queryWords []string) map[int]bool {
	positions := make(map[int]bool, len(words))

	for wordPos, word := range words {
		normalizedWord := pkg.NormalizeString(word)

		for _, qWord := range queryWords {
			if shouldHighlight(normalizedWord, qWord) {
				positions[wordPos] = true
				break
			}
		}
	}

	return positions
}

func shouldHighlight(word, queryWord string) bool {
	return word == queryWord ||
		(len(queryWord) > 3 && strings.Contains(word, queryWord)) ||
		(len(word) > 3 && strings.Contains(queryWord, word))
}
