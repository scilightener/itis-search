package index

import (
	"math"
	"sort"
	"strings"
)

type SearchEngine struct {
	index      *indexData
	tfidf      map[string]map[int64]float64
	docVectors map[int64]map[string]float64
	docNorms   map[int64]float64
}

type SearchResult struct {
	DocID int64
	Score float64
}

func NewSearchEngine(index *indexData) *SearchEngine {
	engine := &SearchEngine{
		index:      index,
		tfidf:      make(map[string]map[int64]float64),
		docVectors: make(map[int64]map[string]float64),
		docNorms:   make(map[int64]float64),
	}
	engine.buildTFIDF()
	engine.precomputeDocNorms()
	return engine
}

func (s *SearchEngine) buildTFIDF() {
	calculator := &tfidfCalculator{index: s.index}
	s.tfidf = calculator.CalculateTFIDF()

	for word, docWeights := range s.tfidf {
		for docID, weight := range docWeights {
			if s.docVectors[docID] == nil {
				s.docVectors[docID] = make(map[string]float64)
			}
			s.docVectors[docID][word] = weight
		}
	}
}

func (s *SearchEngine) precomputeDocNorms() {
	for docID, wordWeights := range s.docVectors {
		var norm float64
		for _, weight := range wordWeights {
			norm += weight * weight
		}
		s.docNorms[docID] = math.Sqrt(norm)
	}
}

func (s *SearchEngine) vectorizeQuery(queryWords []string) map[string]float64 {
	queryVector := make(map[string]float64)
	wordCounts := make(map[string]int)

	for _, word := range queryWords {
		wordCounts[word]++
	}

	for word, count := range wordCounts {
		if _, exists := s.tfidf[word]; !exists {
			continue
		}
		tf := float64(count) / float64(len(queryWords))
		idf := math.Log(float64(s.index.totalDocs) / float64(len(s.index.word2docIDs[word])))
		queryVector[word] = tf * idf
	}

	return queryVector
}

func (s *SearchEngine) cosineSimilarity(queryVector map[string]float64, docID int64) float64 {
	var dotProduct float64
	var queryNorm float64

	for word, queryWeight := range queryVector {
		docWeight := s.docVectors[docID][word]
		dotProduct += queryWeight * docWeight
		queryNorm += queryWeight * queryWeight
	}
	queryNorm = math.Sqrt(queryNorm)

	docNorm := s.docNorms[docID]
	if queryNorm == 0 || docNorm == 0 {
		return 0
	}
	return dotProduct / (queryNorm * docNorm)
}

func (s *SearchEngine) Search(query string, numResults int) []SearchResult {
	queryWords := strings.Fields(strings.ToLower(query))
	queryVector := s.vectorizeQuery(queryWords)

	results := make([]SearchResult, 0)
	for docID := range s.index.docIDs {
		score := s.cosineSimilarity(queryVector, docID)
		if score > 0 {
			results = append(results, SearchResult{DocID: docID, Score: score})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > numResults {
		return results[:numResults]
	}
	return results
}
