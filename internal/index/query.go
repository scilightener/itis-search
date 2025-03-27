package index

import (
	"strings"
	"unicode"

	"search/internal/pkg"
)

type TokenType int

const (
	TokenAnd TokenType = iota + 1
	TokenOr
	TokenNot
	TokenText
)

type Token struct {
	Type TokenType
	Text string
}

type Query struct {
	Tokens []Token
}

func NewQuery(query string) Query {
	q := ParseQuery(query)
	q = NormalizeQuery(q)
	return q
}

func ParseQuery(input string) Query {
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, input)

	var query Query
	var currentText strings.Builder

	for i := 0; i < len(cleaned); i++ {
		char := cleaned[i]

		switch char {
		case '&':
			if currentText.Len() > 0 {
				query.Tokens = append(query.Tokens, Token{Type: TokenText, Text: currentText.String()})
				currentText.Reset()
			}
			query.Tokens = append(query.Tokens, Token{Type: TokenAnd, Text: and})
		case '|':
			if currentText.Len() > 0 {
				query.Tokens = append(query.Tokens, Token{Type: TokenText, Text: currentText.String()})
				currentText.Reset()
			}
			query.Tokens = append(query.Tokens, Token{Type: TokenOr, Text: or})
		case '!':
			if currentText.Len() > 0 {
				query.Tokens = append(query.Tokens, Token{Type: TokenText, Text: currentText.String()})
				currentText.Reset()
			}
			query.Tokens = append(query.Tokens, Token{Type: TokenNot, Text: not})
		default:
			currentText.WriteByte(char)
		}
	}

	if currentText.Len() > 0 {
		query.Tokens = append(query.Tokens, Token{Type: TokenText, Text: currentText.String()})
	}

	return query
}

func NormalizeQuery(query Query) Query {
	for i := range query.Tokens {
		token := &query.Tokens[i]

		if token.Type != TokenText {
			continue
		}

		normalized := pkg.NormalizeString(token.Text)
		if normalized != "" {
			token.Text = normalized
			continue
		}

		shift := 1
		if i > 0 && query.Tokens[i-shift].Type == TokenNot {
			query.Tokens[i-shift].Text = ""
			shift++
		}

		if i > 0 && query.Tokens[i-shift].Type == TokenAnd {
			query.Tokens[i-shift].Text = ""
		}

		if i > 0 && query.Tokens[i-shift].Type == TokenOr {
			if i < len(query.Tokens)-1 {
				query.Tokens[i+1].Text = ""
			} else {
				query.Tokens[i-shift].Text = ""
			}
		}

		query.Tokens[i].Text = ""
	}

	return removeEmptyTokens(query)
}

func removeEmptyTokens(q Query) Query {
	newQuery := Query{
		Tokens: make([]Token, 0, len(q.Tokens)),
	}

	for _, token := range q.Tokens {
		if token.Text != "" {
			newQuery.Tokens = append(newQuery.Tokens, token)
		}
	}

	return newQuery
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
