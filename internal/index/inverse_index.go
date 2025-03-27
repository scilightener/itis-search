package index

const (
	and = "&"
	or  = "|"
	not = "!"
)

type InverseIndex struct {
	indexData
	searchEngine
	tfidfCalculator
}

func NewInverseIndex() *InverseIndex {
	data := &indexData{
		word2docIDs:    make(map[string]map[int64]struct{}),
		word2docCounts: make(map[string]map[int64]int),
		docLengths:     make(map[int64]int),
		docIDs:         make(map[int64]struct{}),
		totalDocs:      0,
	}
	return &InverseIndex{
		indexData: *data,
		searchEngine: searchEngine{
			data,
		},
		tfidfCalculator: tfidfCalculator{
			data,
		},
	}
}

func (i *InverseIndex) Load(indexPath string) error {
	err := i.indexData.Load(indexPath)
	if err != nil {
		return err
	}

	i.searchEngine.index = &i.indexData
	i.tfidfCalculator.index = &i.indexData
	return nil
}

func setToSlice(set map[int64]struct{}) []int64 {
	res := make([]int64, 0, len(set))
	for k := range set {
		res = append(res, k)
	}

	return res
}
