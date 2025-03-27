package task

import "context"

func NotFinishedFilter(_ context.Context, t *Task) bool {
	return !t.Finished
}
