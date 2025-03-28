package index

import (
	"encoding/json"
	"os"
	"slices"
	"strings"

	"search/internal/domain"
)

type indexData struct {
	word2docIDs    map[string]map[int64]struct{}
	word2docCounts map[string]map[int64]int
	docLengths     map[int64]int
	docIDs         map[int64]struct{}
	totalDocs      int
}

func (d *indexData) Add(doc domain.Document) {
	words := strings.Fields(string(doc.ProcessedText))
	if _, exists := d.docIDs[doc.ID]; !exists {
		d.totalDocs++
	}
	d.docIDs[doc.ID] = struct{}{}
	d.docLengths[doc.ID] = len(words)

	wordCounts := make(map[string]int)
	for _, word := range words {
		wordCounts[word]++
	}

	for word, count := range wordCounts {
		if _, ok := d.word2docIDs[word]; !ok {
			d.word2docIDs[word] = make(map[int64]struct{})
			d.word2docCounts[word] = make(map[int64]int)
		}
		d.word2docIDs[word][doc.ID] = struct{}{}
		d.word2docCounts[word][doc.ID] = count
	}
}

func (d *indexData) Save(filename string) error {
	word2DocIDs := make(map[string][]int64)
	for word, docIDs := range d.word2docIDs {
		ids := setToSlice(docIDs)
		slices.Sort(ids)
		word2DocIDs[word] = ids
	}
	data := struct {
		Word2DocIDs    map[string][]int64       `json:"word2doc_ids"`
		Word2DocCounts map[string]map[int64]int `json:"word2doc_counts"`
		DocLengths     map[int64]int            `json:"doc_lengths"`
		TotalDocs      int                      `json:"total_docs"`
	}{
		Word2DocIDs:    word2DocIDs,
		Word2DocCounts: d.word2docCounts,
		DocLengths:     d.docLengths,
		TotalDocs:      d.totalDocs,
	}

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

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (d *indexData) Load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	var data struct {
		Word2DocIDs    map[string][]int64       `json:"word2doc_ids"`
		Word2DocCounts map[string]map[int64]int `json:"word2doc_counts"`
		DocLengths     map[int64]int            `json:"doc_lengths"`
		TotalDocs      int                      `json:"total_docs"`
	}

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return err
	}

	d.word2docIDs = make(map[string]map[int64]struct{})
	for word, docIDs := range data.Word2DocIDs {
		d.word2docIDs[word] = make(map[int64]struct{})
		for _, id := range docIDs {
			d.word2docIDs[word][id] = struct{}{}
		}
	}

	d.word2docCounts = data.Word2DocCounts
	d.docLengths = data.DocLengths
	d.totalDocs = data.TotalDocs

	d.docIDs = make(map[int64]struct{})
	for id := range d.docLengths {
		d.docIDs[id] = struct{}{}
	}

	return nil
}
