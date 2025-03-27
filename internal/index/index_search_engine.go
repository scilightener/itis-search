package index

type searchEngine struct {
	index *indexData
}

func (s *searchEngine) Search(query Query) []int64 {
	if containsOperator(query, TokenOr) {
		return s.handleOr(query)
	}

	if containsOperator(query, TokenAnd) {
		return s.handleAnd(query)
	}

	return s.handleBaseCase(query)
}

func (s *searchEngine) handleBaseCase(query Query) []int64 {
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

	docIDs, exists := s.index.word2docIDs[textToken.Text]
	if !exists {
		docIDs = make(map[int64]struct{})
	}

	if negate {
		result := make(map[int64]struct{})
		for docID := range s.index.docIDs {
			if _, found := docIDs[docID]; !found {
				result[docID] = struct{}{}
			}
		}
		return setToSlice(result)
	}

	return setToSlice(docIDs)
}

func (s *searchEngine) handleOr(query Query) []int64 {
	subqueries := splitQuery(query, TokenOr)
	result := make(map[int64]struct{})

	for _, subquery := range subqueries {
		docIDs := s.Search(subquery)
		for _, docID := range docIDs {
			result[docID] = struct{}{}
		}
	}

	return setToSlice(result)
}

func (s *searchEngine) handleAnd(query Query) []int64 {
	subqueries := splitQuery(query, TokenAnd)
	var result map[int64]struct{}

	for _, subquery := range subqueries {
		docIDs := s.Search(subquery)
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
