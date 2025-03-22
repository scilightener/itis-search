package pipe

import (
	"context"
)

// Pipe is a function that takes a channel, processes its values,
// and returns a new channel with the processed values from the old one.
// Note: several Pipe instances can be run at the same time,
// so it should synchronize itself.
type Pipe[T any] func(ctx context.Context, in <-chan T) <-chan T

// Source is a function that sends values in the output channel.
type Source[T any] func(ctx context.Context) <-chan T

type Handler[T any] func(ctx context.Context, t T) T

func NewPipe[T any](h Handler[T]) Pipe[T] {
	return func(ctx context.Context, in <-chan T) <-chan T {
		out := make(chan T, cap(in))

		go func() {
			defer close(out)

			for t := range in {
				if err := ctx.Err(); err != nil {
					break
				}

				t = h(ctx, t)

				out <- t
			}
		}()

		return out
	}
}

// StartPipeline links all the pipelines into one big pipeline and starts Source.
func StartPipeline[T any](ctx context.Context, start Source[T], pipes ...Pipe[T]) <-chan T {
	out := start(ctx)
	for _, p := range pipes {
		out = p(ctx, out)
	}

	return out
}
