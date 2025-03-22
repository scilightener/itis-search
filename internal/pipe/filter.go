package pipe

import "context"

type Filter[T any] func(ctx context.Context, t T) bool

func Satisfies[T any](satisfy Filter[T]) Pipe[T] {
	return func(ctx context.Context, in <-chan T) <-chan T {
		out := make(chan T, cap(in))

		go func() {
			defer close(out)

			for t := range in {
				if err := ctx.Err(); err != nil {
					break
				}

				if !satisfy(ctx, t) {
					continue
				}

				out <- t
			}
		}()

		return out
	}
}
