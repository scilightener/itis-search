package task

import (
	"context"
	"net/url"
	"strings"
	"sync/atomic"

	"search/internal/pipe"
	"search/internal/pkg"
)

var nextID atomic.Int64

// NewTaskGenerator returns a new Source that generates Task's.
func NewTaskGenerator(chanCapacity int, linksChan <-chan string, stopChan <-chan struct{}) pipe.Source[*Task] {
	return func(ctx context.Context) <-chan *Task {
		out := make(chan *Task, chanCapacity)

		seen := make(map[string]bool)

		go func() {
			defer close(out)
			for link := range linksChan {
				select {
				case <-ctx.Done():
					return
				case <-stopChan:
					return
				default:
					if !isValidLink(link) {
						continue
					}

					link = pkg.PrepareLink(link)
					if seen[link] {
						continue
					}

					seen[link] = true
					id := nextID.Add(1)

					out <- NewTask(id, link)
				}
			}
		}()

		return out
	}
}

func isValidLink(link string) bool {
	link = strings.TrimSpace(link)
	if link == "" {
		return false
	}

	parsedURL, err := url.Parse(link)
	return err == nil && parsedURL.Scheme != "" && parsedURL.Host != ""
}
