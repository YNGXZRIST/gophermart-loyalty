package workerpool

import (
	"context"
)

type Worker struct {
	tCh    chan Task
	rCh    chan Task
	quitCh chan any
}

func NewWorker(tCh, rCh chan Task) *Worker {
	return &Worker{tCh: tCh, rCh: rCh}
}

func (w *Worker) StartBg(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case t, ok := <-w.tCh:
			if !ok {
				return
			}
			t.process()
			if t.NeedResult {
				select {
				case w.rCh <- t:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}
