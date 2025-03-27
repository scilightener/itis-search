package index

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

type tfidfCalculator struct {
	index *indexData
}

func (t *tfidfCalculator) CalculateTF() map[string]map[int64]float64 {
	tf := make(map[string]map[int64]float64)

	for word, docCounts := range t.index.word2docCounts {
		tf[word] = make(map[int64]float64)
		for docID, count := range docCounts {
			tf[word][docID] = float64(count) / float64(t.index.docLengths[docID])
		}
	}

	return tf
}

func (t *tfidfCalculator) CalculateIDF() map[string]float64 {
	idf := make(map[string]float64)

	for word, docIDs := range t.index.word2docIDs {
		docCount := len(docIDs)
		idf[word] = math.Log(float64(t.index.totalDocs) / float64(docCount))
	}

	return idf
}

func (t *tfidfCalculator) CalculateTFIDF() map[string]map[int64]float64 {
	tf := t.CalculateTF()
	idf := t.CalculateIDF()
	tfidf := make(map[string]map[int64]float64)

	for word, docTFs := range tf {
		tfidf[word] = make(map[int64]float64)
		for docID, tfValue := range docTFs {
			tfidf[word][docID] = tfValue * idf[word]
		}
	}

	return tfidf
}

func SaveTable(filename, header string, data map[string]map[int64]float64) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	writer := csv.NewWriter(file)
	defer writer.Flush()

	titles := strings.FieldsFunc(header, func(r rune) bool {
		return r == ',' || r == '\t' || r == ' ' || r == '|'
	})
	if err := writer.Write(titles); err != nil {
		return err
	}

	words := make([]string, 0, len(data))
	for word := range data {
		words = append(words, word)
	}
	slices.Sort(words)

	for _, word := range words {
		docValues := data[word]

		docIDs := make([]int64, 0, len(docValues))
		for id := range docValues {
			docIDs = append(docIDs, id)
		}
		slices.Sort(docIDs)

		for _, id := range docIDs {
			value := fmt.Sprintf("%.6f", docValues[id])
			if err := writer.Write([]string{word, strconv.FormatInt(id, 10), value}); err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *tfidfCalculator) SaveTF(filename string) error {
	tf := t.CalculateTF()
	return SaveTable(filename, "word,doc_id,tf", tf)
}

func (t *tfidfCalculator) SaveIDF(filename string) error {
	idf := t.CalculateIDF()
	data := make(map[string]map[int64]float64)
	for word, value := range idf {
		data[word] = map[int64]float64{0: value}
	}
	return SaveTable(filename, "word,doc_id,idf", data)
}

func (t *tfidfCalculator) SaveTFIDF(filename string) error {
	tfidf := t.CalculateTFIDF()
	return SaveTable(filename, "word,doc_id,tfidf", tfidf)
}

func PrintTable(title string, data map[string]map[int64]float64, maxRows int) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Word", "Document ID", "Value"})
	table.SetBorder(true)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetColumnSeparator("│")
	table.SetRowSeparator("─")
	table.SetHeaderLine(true)
	table.SetCaption(true, title)

	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiGreenColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiYellowColor},
	)

	words := make([]string, 0, len(data))
	for word := range data {
		words = append(words, word)
	}
	slices.Sort(words)

	if maxRows > 0 && len(words) > maxRows {
		words = words[:maxRows]
	}

	for _, word := range words {
		docValues := data[word]

		docIDs := make([]int64, 0, len(docValues))
		for docID := range docValues {
			docIDs = append(docIDs, docID)
		}
		slices.Sort(docIDs)

		for _, docID := range docIDs {
			value := docValues[docID]
			table.Append([]string{
				word,
				strconv.FormatInt(docID, 10),
				fmt.Sprintf("%.6f", value),
			})
		}
	}

	table.Render()
}
