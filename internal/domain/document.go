package domain

import "fmt"

type Document struct {
	Text  []byte
	Links []string
	URI   string
}

func NewDocument(uri string) Document {
	return Document{
		Text:  make([]byte, 0),
		Links: make([]string, 0),
		URI:   uri,
	}
}

func (d Document) String() string {
	return fmt.Sprintf("Document: {URI: %s, Links length: %d, Text length: %d}",
		d.URI, len(d.Links), len(d.Text))
}

func (d Document) Clone() Document {
	clone := Document{
		Text:  make([]byte, len(d.Text)),
		Links: make([]string, len(d.Links)),
		URI:   d.URI,
	}

	copy(clone.Text, d.Text)
	copy(clone.Links, d.Links)

	return clone
}
