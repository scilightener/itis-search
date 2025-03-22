package pipe

import "context"

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
