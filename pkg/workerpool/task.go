// Package workerpool provides a lightweight in-memory worker pool.
package workerpool

// Task is a unit of work processed by worker pool.
type Task struct {
	Err        error
	Result     any
	f          func(any) (any, error)
	NeedResult bool
}

func (t *Task) process() {
	t.Result, t.Err = t.f(t.Result)
}

// NewTask creates task from processing function.
func NewTask(f func(any) (any, error)) *Task {
	return &Task{
		f:          f,
		NeedResult: true,
	}
}
