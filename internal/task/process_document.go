package task

import (
	"context"
	"regexp"
	"strings"

	"github.com/bbalet/stopwords"
	"github.com/kljensen/snowball"

	"search/internal/domain"
)

func ProcessDocumentHandler(_ context.Context, t *Task) *Task {
	t.Document = processDocument(t.Document)
	return t
}

func processDocument(doc domain.Document) domain.Document {
	words := tokenize(doc.Text)
	words = removeStopWords(words)
	words = lemmatize(words)

	doc.Text = []byte(strings.Join(words, " "))
	return doc
}

func tokenize(text []byte) []string {
	re := regexp.MustCompile(`[а-яА-ЯёЁ]+`)
	words := re.FindAllString(strings.ToLower(string(text)), -1)
	return words
}

func lemmatize(words []string) []string {
	lemmatizedWords := make([]string, 0, len(words))
	for _, word := range words {
		stemmed, err := snowball.Stem(word, "russian", true)
		if err == nil {
			lemmatizedWords = append(lemmatizedWords, stemmed)
		}
	}
	return lemmatizedWords
}

func removeStopWords(words []string) []string {
	cleanedText := stopwords.CleanString(strings.Join(words, " "), "ru", true)
	return strings.Fields(cleanedText)
}
