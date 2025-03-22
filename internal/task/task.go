package task

import (
	"fmt"
	"time"

	"search/internal/domain"
	"search/internal/result"
)

type Task struct {
	ID           int64
	Link         string
	Document     domain.Document
	CreationTime time.Time
	Finished     bool
	FinishTime   time.Time
	Result       result.Result
}

func NewTask(id int64, link string) *Task {
	return &Task{
		ID:           id,
		Link:         link,
		CreationTime: time.Now(),
		Finished:     false,
		FinishTime:   time.Time{},
		Result:       result.NewResult(true).WithMessages("just created"),
		Document:     domain.NewDocument(link),
	}
}

// Finish finishes the task with provided Result.
// It returns old task, but finished.
func (t *Task) Finish(result result.Result) *Task {
	t.Result = result
	t.Finished = true
	t.FinishTime = time.Now()
	return t
}

// Fail fails the task with the provided messages
// It returns old task, but finished.
func (t *Task) Fail(messages ...string) *Task {
	if t.Finished {
		return t
	}

	res := result.NewResult(false).WithMessages(messages...)
	return t.Finish(res)
}

// String returns string representation of a Task.
func (t *Task) String() string {
	return fmt.Sprintf("Task: {id: %d, creation time: %s, finish time: %s, finished: %t, link: %s, document: %v, result: %s}",
		t.ID,
		t.CreationTime.Format(time.RFC3339Nano),
		t.FinishTime.Format(time.RFC3339Nano),
		t.Finished,
		t.Link,
		t.Document.String(),
		t.Result.String(),
	)
}
