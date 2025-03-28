package domain

import "fmt"

type Document struct {
	ID            int64
	RawText       []byte
	ProcessedText []byte
	Links         []string
	URI           string
}

func NewDocument(id int64, uri string) Document {
	return Document{
		ID:            id,
		RawText:       make([]byte, 0),
		ProcessedText: make([]byte, 0),
		Links:         make([]string, 0),
		URI:           uri,
	}
}

func (d Document) String() string {
	return fmt.Sprintf("Document: {ID: %d, URI: %s, Links length: %d, ProcessedText length: %d}",
		d.ID, d.URI, len(d.Links), len(d.ProcessedText))
}
