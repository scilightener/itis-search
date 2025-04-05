package index

type Index struct {
	Data indexData
	tfidfCalculator
}

func NewIndex() *Index {
	data := &indexData{
		word2docIDs:    make(map[string]map[int64]struct{}),
		word2docCounts: make(map[string]map[int64]int),
		docLengths:     make(map[int64]int),
		docID2link:     make(map[int64]string),
		totalDocs:      0,
	}
	return &Index{
		Data: *data,
		tfidfCalculator: tfidfCalculator{
			data,
		},
	}
}

func (i *Index) Load(indexPath string) error {
	err := i.Data.Load(indexPath)
	if err != nil {
		return err
	}

	i.tfidfCalculator.index = &i.Data
	return nil
}

func setToSlice(set map[int64]struct{}) []int64 {
	res := make([]int64, 0, len(set))
	for k := range set {
		res = append(res, k)
	}

	return res
}
