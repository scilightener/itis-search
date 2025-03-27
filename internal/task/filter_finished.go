package task

import "context"

func FinishedFilter(_ context.Context, t *Task) bool {
	return !t.Finished
}
