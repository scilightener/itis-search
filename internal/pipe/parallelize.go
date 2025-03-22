package pipe

import (
	"context"
	"sync"
)

// Parallelize is a Pipe that runs the incoming Pipe in parallel of numWorkers goroutines.
func Parallelize[T any](p Pipe[T], numWorkers int) Pipe[T] {
	return func(ctx context.Context, in <-chan T) <-chan T {
		out := make(chan T, cap(in))
		var wg sync.WaitGroup

		wg.Add(numWorkers)
		for i := 0; i < numWorkers; i++ {
			go func() {
				defer wg.Done()
				for v := range p(ctx, in) {
					select {
					case <-ctx.Done():
						return
					default:
						out <- v
					}
				}
			}()
		}

		go func() {
			wg.Wait()
			close(out)
		}()

		return out
	}
}
