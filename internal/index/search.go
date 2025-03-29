package index

import "slices"

type Search struct {
	index *InverseIndex
}

func NewSearch(index *InverseIndex) *Search {
	return &Search{
		index: index,
	}
}

func (s *Search) Search(query string) []int64 {
	q := NewQuery(query)
	res := s.index.Search(q)
	slices.Sort(res)
	return res
}
