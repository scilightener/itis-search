package index

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"

	"search/internal/pkg"
)

type SearchEngine struct {
	index      *indexData
	tfidf      map[string]map[int64]float64
	docVectors map[int64]map[string]float64
	docNorms   map[int64]float64
	docCache   map[int64]string
	cacheMutex sync.RWMutex
}

type SearchResult struct {
	DocID   int64
	Score   float64
	Snippet string
}

func NewSearchEngine(index *indexData) *SearchEngine {
	engine := &SearchEngine{
		index:      index,
		docCache:   make(map[int64]string),
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

func (s *SearchEngine) Search(query string, numResults, windowSize int) []SearchResult {
	query = pkg.NormalizeString(query)
	queryWords := strings.Fields(query)
	queryVector := s.vectorizeQuery(queryWords)

	type docScore struct {
		docID int64
		score float64
	}

	var (
		mu        sync.Mutex
		wg        sync.WaitGroup
		docScores = make([]docScore, 0)
	)

	for docID := range s.index.docIDs {
		wg.Add(1)
		go func(docID int64) {
			defer wg.Done()
			score := s.cosineSimilarity(queryVector, docID)
			if score > 0 {
				mu.Lock()
				docScores = append(docScores, docScore{docID, score})
				mu.Unlock()
			}
		}(docID)
	}

	wg.Wait()

	sort.Slice(docScores, func(i, j int) bool {
		return docScores[i].score > docScores[j].score
	})

	topDocs := docScores
	if len(topDocs) > numResults {
		topDocs = topDocs[:numResults]
	}

	results := make([]SearchResult, len(topDocs))
	for i, ds := range topDocs {
		snippet := s.findBestSnippet(ds.docID, queryWords, windowSize)
		results[i] = SearchResult{
			DocID:   ds.docID,
			Score:   ds.score,
			Snippet: snippet,
		}
	}

	return results
}

func (s *SearchEngine) findBestSnippet(docID int64, queryWords []string, windowSize int) string {
	content, err := s.getDocumentContent(docID)
	if err != nil || content == "" {
		return ""
	}

	words := strings.Fields(content)
	totalWords := len(words)
	if totalWords == 0 {
		return ""
	}

	if windowSize > totalWords {
		windowSize = totalWords
	}

	var (
		bestScore int
		bestStart int
		mu        sync.Mutex
	)

	numWorkers := runtime.NumCPU()
	chunkSize := totalWords / numWorkers
	if chunkSize < windowSize {
		chunkSize = windowSize
	}

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		start := i * chunkSize
		end := start + chunkSize + windowSize
		if end > totalWords {
			end = totalWords
		}
		if start >= end {
			continue
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()

			localBestScore := -1
			localBestStart := 0

			for i := start; i <= end-windowSize/2; i++ {
				score := 0
				for j := i; j < i+windowSize/2 && j < totalWords; j++ {
					word := pkg.NormalizeString(words[j])
					for _, qWord := range queryWords {
						if word == qWord {
							score++
							break
						}
					}
				}

				if score >= localBestScore {
					localBestScore = score
					localBestStart = i
				}
			}

			mu.Lock()
			if localBestScore > bestScore ||
				(localBestScore == bestScore && localBestStart < bestStart) {
				bestScore = localBestScore
				bestStart = localBestStart
			}
			mu.Unlock()
		}(start, end)
	}

	wg.Wait()

	snippetEnd := bestStart + windowSize/2
	if snippetEnd > totalWords {
		snippetEnd = totalWords
	}

	snippetStart := bestStart - windowSize/2
	if snippetStart < 0 {
		snippetStart = 0
	}

	return strings.Join(words[snippetStart:snippetEnd], " ")
}

const docDirPath = "data/5/raw"

func (s *SearchEngine) getDocumentContent(docID int64) (string, error) {
	s.cacheMutex.RLock()
	content, ok := s.docCache[docID]
	s.cacheMutex.RUnlock()

	if ok {
		return content, nil
	}

	filePath := fmt.Sprintf("%s/%d.txt", docDirPath, docID)
	rawContent, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	content = string(rawContent)

	s.cacheMutex.Lock()
	s.docCache[docID] = content
	s.cacheMutex.Unlock()

	return content, nil
}
