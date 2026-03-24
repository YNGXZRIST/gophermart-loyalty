package workerpool

import (
	"context"
)

// Worker consumes tasks and optionally publishes results.
type Worker struct {
	tCh    chan Task
	rCh    chan Task
	quitCh chan any
}

// NewWorker creates worker bound to task and result channels.
func NewWorker(tCh, rCh chan Task) *Worker {
	return &Worker{tCh: tCh, rCh: rCh}
}

// StartBg starts worker loop and exits on context cancellation.
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
