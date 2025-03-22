package task

import (
	"context"
	"fmt"
	"sync/atomic"

	"search/internal/pipe"
)

func NewLogAsyncHandler() pipe.AsyncHandler[*Task] {
	var n atomic.Int64

	return func(ctx context.Context, t *Task) {
		if t.Finished {
			return
		}

		curIdx := n.Add(1)
		fmt.Printf("%d %s\n", curIdx, t.String())
	}
}
