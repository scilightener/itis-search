package task

import (
	"context"
	"fmt"
	"sync/atomic"

	"search/internal/pipe"
)

func NewLogAsyncPipe() pipe.AsyncPipe[*Task] {
	return func(ctx context.Context, in <-chan *Task) {
		var n atomic.Int64

		go func() {
			for t := range in {
				if err := ctx.Err(); err != nil {
					break
				}

				if t.Finished {
					continue
				}

				curIdx := n.Add(1)
				fmt.Printf("%d %s\n", curIdx, t.String())
			}
		}()
	}
}
