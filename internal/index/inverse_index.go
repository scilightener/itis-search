package index

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"

	"search/internal/domain"
)

const (
	and = "&"
	or  = "|"
	not = "!"
)

type InverseIndex struct {
	word2docIDs map[string]map[int64]struct{}
	docIDs      map[int64]struct{}
}

func NewInverseIndex() *InverseIndex {
	return &InverseIndex{
		word2docIDs: make(map[string]map[int64]struct{}),
		docIDs:      make(map[int64]struct{}),
	}
}

func (i *InverseIndex) Add(d domain.Document) {
	words := strings.Fields(string(d.Text))
	for _, word := range words {
		if i.word2docIDs[word] == nil {
			i.word2docIDs[word] = make(map[int64]struct{})
		}
		i.word2docIDs[word][d.ID] = struct{}{}
	}
	i.docIDs[d.ID] = struct{}{}
}

func (i *InverseIndex) Save(fileName string) error {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0774)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	words := make([]string, 0, len(i.word2docIDs))
	for word := range i.word2docIDs {
		words = append(words, word)
	}
	sort.Strings(words)

	for _, word := range words {
		docIDs := setToSlice(i.word2docIDs[word])
		slices.Sort(docIDs)
		_, err := fmt.Fprintf(file, "%s %v\n", word, docIDs)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *InverseIndex) Load(fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}

		word := parts[0]
		docIDsStr := parts[1]

		docIDsStr = strings.Trim(docIDsStr, "[]")
		if docIDsStr == "" {
			continue
		}

		idStrs := strings.Split(docIDsStr, " ")
		var docIDs []int64
		for _, idStr := range idStrs {
			if idStr == "" {
				continue
			}
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse docID '%s': %v", idStr, err)
			}
			docIDs = append(docIDs, id)
			i.docIDs[id] = struct{}{}
		}

		i.word2docIDs[word] = sliceToSet(docIDs)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	return nil
}

func (i *InverseIndex) Search(query Query) []int64 {
	if containsOperator(query, TokenOr) {
		return i.handleOr(query)
	}

	if containsOperator(query, TokenAnd) {
		return i.handleAnd(query)
	}

	return i.handleBaseCase(query)
}

func (i *InverseIndex) handleBaseCase(query Query) []int64 {
	if len(query.Tokens) < 1 {
		return nil
	}

	var textToken *Token
	negate := false

	if query.Tokens[0].Type == TokenNot {
		negate = true
		if len(query.Tokens) > 1 && query.Tokens[1].Type == TokenText {
			textToken = &query.Tokens[1]
		} else {
			return nil
		}
	} else if query.Tokens[0].Type == TokenText {
		textToken = &query.Tokens[0]
	} else {
		return nil
	}

	docIDs, exists := i.word2docIDs[textToken.Text]
	if !exists {
		docIDs = make(map[int64]struct{})
	}

	if negate {
		result := make(map[int64]struct{})
		for docID := range i.docIDs {
			if _, found := docIDs[docID]; !found {
				result[docID] = struct{}{}
			}
		}
		return setToSlice(result)
	}

	return setToSlice(docIDs)
}

func (i *InverseIndex) handleOr(query Query) []int64 {
	subqueries := splitQuery(query, TokenOr)
	result := make(map[int64]struct{})

	for _, subquery := range subqueries {
		docIDs := i.Search(subquery)
		for _, docID := range docIDs {
			result[docID] = struct{}{}
		}
	}

	return setToSlice(result)
}

func (i *InverseIndex) handleAnd(query Query) []int64 {
	subqueries := splitQuery(query, TokenAnd)
	var result map[int64]struct{}

	for _, subquery := range subqueries {
		docIDs := i.Search(subquery)
		current := make(map[int64]struct{})
		for _, docID := range docIDs {
			current[docID] = struct{}{}
		}

		if result == nil {
			result = current
		} else {
			for docID := range result {
				if _, exists := current[docID]; !exists {
					delete(result, docID)
				}
			}
		}
	}

	return setToSlice(result)
}

func splitQuery(query Query, op TokenType) []Query {
	var subqueries []Query
	start := 0

	for i, token := range query.Tokens {
		if token.Type == op {
			if i > start {
				subqueries = append(subqueries, Query{Tokens: query.Tokens[start:i]})
			}
			start = i + 1
		}
	}

	if start < len(query.Tokens) {
		subqueries = append(subqueries, Query{Tokens: query.Tokens[start:]})
	}

	return subqueries
}

func containsOperator(query Query, op TokenType) bool {
	for _, token := range query.Tokens {
		if token.Type == op {
			return true
		}
	}
	return false
}

func setToSlice(set map[int64]struct{}) []int64 {
	res := make([]int64, 0, len(set))
	for k := range set {
		res = append(res, k)
	}

	return res
}

func sliceToSet(slice []int64) map[int64]struct{} {
	res := make(map[int64]struct{})
	for _, id := range slice {
		res[id] = struct{}{}
	}
	return res
}
