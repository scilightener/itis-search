package pipe

import "context"

// Pipe is a function that takes a channel, processes its values,
// and returns a new channel with the processed values from the old one.
// Note: several Pipe instances can be run at the same time,
// so it should synchronize itself.
type Pipe[T any] func(ctx context.Context, in <-chan T) <-chan T

// AsyncPipe is a function that takes a channel, processes its values,
// and returns a new channel with the processed values from the old one.
// Note: several AsyncPipe instances can be run at the same time,
// so it should synchronize itself.
type AsyncPipe[T any] func(ctx context.Context, in <-chan T)

// Source is a function that sends values in the output channel.
type Source[T any] func(ctx context.Context) <-chan T

// StartPipeline links all the pipelines into one big pipeline and starts Source.
func StartPipeline[T any](ctx context.Context, start Source[T], pipes ...Pipe[T]) <-chan T {
	out := start(ctx)
	for _, p := range pipes {
		out = p(ctx, out)
	}

	return out
}

// NewDiscardPipe is a Pipe for discarding everything that passes into its input.
// It returns nil output channel.
func NewDiscardPipe[T any]() Pipe[T] {
	return func(ctx context.Context, in <-chan T) <-chan T {
		go func() {
			for range in {
				if err := ctx.Err(); err != nil {
					break
				}
			}
		}()

		return nil
	}
}

// NewParallelPipe is a Pipe that runs the incoming Pipe in parallel of numWorkers goroutines.
func NewParallelPipe[T any](p Pipe[T], numWorkers int) Pipe[T] {
	return func(ctx context.Context, in <-chan T) <-chan T {
		out := make(chan T, cap(in))
		stop := make(chan struct{})

		for i := 0; i < numWorkers; i++ {
			go func() {
				for v := range p(ctx, in) {
					select {
					case <-ctx.Done():
						stop <- struct{}{}
						return
					case <-stop:
						return
					default:
						out <- v
					}
				}
			}()
		}

		go func() {
			<-stop
			close(out)
		}()

		return out
	}
}

func NewPipeFromAsyncPipe[T any](asyncPipe AsyncPipe[T]) Pipe[T] {
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
