package pipe

import (
	"context"
)

func NewWaitPipe[T any]() Pipe[T] {
	return func(ctx context.Context, in <-chan T) <-chan T {
		out := make(chan T)

		go func() {
			defer close(out)
			defer func() {
				var t T
				out <- t
			}()

			for range in {
				if err := ctx.Err(); err != nil {
					break
				}
			}
		}()

		return out
	}
}
