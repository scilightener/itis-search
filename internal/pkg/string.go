package pkg

import (
	"regexp"
	"strings"

	"github.com/bbalet/stopwords"
	"github.com/kljensen/snowball"
)

func NormalizeString(s string) string {
	words := tokenize(s)
	words = removeStopWords(words)
	words = lemmatize(words)

	return strings.Join(words, " ")
}

func tokenize(text string) []string {
	re := regexp.MustCompile(`[а-яА-ЯёЁ]+`)
	words := re.FindAllString(strings.ToLower(text), -1)
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
