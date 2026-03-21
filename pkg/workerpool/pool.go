package workerpool

import (
	"context"
)

// Pool manages workers and task/result channels.
type Pool struct {
	workers []*Worker
	rCh     chan Task
	tCh     chan Task
	tRes    *[]Task
	cancel  func()
}

// NewPool creates worker pool with given capacity.
func NewPool(c int) *Pool {
	w := &Pool{
		workers: make([]*Worker, 0, c),
		rCh:     make(chan Task, c),
		tCh:     make(chan Task, c),
		tRes:    &[]Task{},
	}
	return w
}

// StartBg starts all workers in background.
func (p *Pool) StartBg(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	for i := 0; i < cap(p.workers); i++ {
		worker := NewWorker(p.tCh, p.rCh)
		p.workers = append(p.workers, worker)
		go worker.StartBg(ctx)
	}

}

// Add enqueues task for processing.
func (p *Pool) Add(task *Task) {
	p.tCh <- *task
}

// Get blocks until processed task result is available.
func (p *Pool) Get() Task {
	return <-p.rCh
}

// Stop stops workers by canceling shared context.
func (p *Pool) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
}
