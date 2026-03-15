package workerpool

import (
	"context"
)

type Worker struct {
	ctx context.Context
	tCh chan Task
	rCh chan Task
}

func NewWorker(ctx context.Context, tCh, rCh chan Task) *Worker {
	return &Worker{ctx: ctx, tCh: tCh, rCh: rCh}
}

func (w *Worker) StartBg() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case t, ok := <-w.tCh:
			if !ok {
				return
			}
			t.process()
			if t.NeedResult {
				w.rCh <- t
			}
		}
	}
}
