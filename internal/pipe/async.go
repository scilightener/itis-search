package pipe

import "context"

// AsyncPipe is a function that takes a channel, processes its values,
// and returns a new channel with the processed values from the old one.
// Note: several AsyncPipe instances can be run at the same time,
// so it should synchronize itself.
type AsyncPipe[T any] func(ctx context.Context, in <-chan T)

type AsyncHandler[T any] func(ctx context.Context, t T)

func NewAsyncPipe[T any](h AsyncHandler[T]) AsyncPipe[T] {
	return func(ctx context.Context, in <-chan T) {
		go func() {
			for t := range in {
				if err := ctx.Err(); err != nil {
					break
				}

				h(ctx, t)
			}
		}()
	}
}

// Synchronize is a function that connects AsyncPipe to the main pipeline.
func Synchronize[T any](asyncPipe AsyncPipe[T]) Pipe[T] {
	return func(ctx context.Context, in <-chan T) <-chan T {
		out := make(chan T, cap(in))
		duplicateOut := make(chan T, cap(in))

		go func() {
			defer close(out)
			defer close(duplicateOut)
			asyncPipe(ctx, duplicateOut)
			for v := range in {
				if err := ctx.Err(); err != nil {
					return
				}
				out <- v
				duplicateOut <- v
			}
		}()

		return out
	}
}
