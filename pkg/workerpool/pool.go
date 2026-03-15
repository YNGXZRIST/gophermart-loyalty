package workerpool

import (
	"context"
)

type Pool struct {
	workers []*Worker
	rCh     chan Task
	tCh     chan Task
	tRes    *[]Task
}

func NewPool(c int) *Pool {
	w := &Pool{
		workers: make([]*Worker, 0, c),
		rCh:     make(chan Task, c),
		tCh:     make(chan Task, c),
		tRes:    &[]Task{},
	}
	return w
}
func (p *Pool) StartBg(ctx context.Context) {
	for i := 0; i < cap(p.workers); i++ {
		worker := NewWorker(ctx, p.tCh, p.rCh)
		p.workers = append(p.workers, worker)
		go worker.StartBg()
	}

}
func (p *Pool) Add(task *Task) {
	p.tCh <- *task
}
func (p *Pool) Get() Task {
	return <-p.rCh
}
