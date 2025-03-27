package domain

import "fmt"

type Document struct {
	ID    int64
	Text  []byte
	Links []string
	URI   string
}

func NewDocument(id int64, uri string) Document {
	return Document{
		ID:    id,
		Text:  make([]byte, 0),
		Links: make([]string, 0),
		URI:   uri,
	}
}

func (d Document) String() string {
	return fmt.Sprintf("Document: {ID: %d, URI: %s, Links length: %d, Text length: %d}",
		d.ID, d.URI, len(d.Links), len(d.Text))
}
