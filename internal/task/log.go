package task

import (
	"context"
	"fmt"

	"search/internal/pipe"
)

func NewLogAsyncPipe() pipe.AsyncPipe[*Task] {
	return func(ctx context.Context, in <-chan *Task) {
		go func() {
			for t := range in {
				if err := ctx.Err(); err != nil {
					break
				}

				fmt.Println(t.String())
			}
		}()
	}
}
