package task

import (
	"context"
	"os"

	"search/internal/index"
	"search/internal/pipe"
)

func NewIndexerPipe(indexFileName string) pipe.AsyncPipe[*Task] {
	err := deleteIndexIfExists(indexFileName)
	if err != nil {
		panic(err)
	}

	idx := index.NewInverseIndex()

	return func(ctx context.Context, in <-chan *Task) {
		go func() {
			defer func(index *index.InverseIndex, fileName string) {
				err := index.Save(fileName)
				if err != nil {
					panic(err)
				}
			}(idx, indexFileName)

			for t := range in {
				if err := ctx.Err(); err != nil {
					break
				}

				if t.Finished {
					continue
				}

				idx.Add(t.Document)
			}
		}()
	}
}

func deleteIndexIfExists(indexFileName string) error {
	if _, err := os.Stat(indexFileName); err == nil {
		if err := os.Remove(indexFileName); err != nil {
			return err
		}
	}

	return nil
}
