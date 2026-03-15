package workerpool

type Task struct {
	Err        error
	Result     any
	f          func(any) (any, error)
	NeedResult bool
}

func (t *Task) process() {
	t.Result, t.Err = t.f(t.Result)
}
func NewTask(f func(any) (any, error)) *Task {
	return &Task{
		f:          f,
		NeedResult: true,
	}
}
