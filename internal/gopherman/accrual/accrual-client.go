package accrual

import (
	"context"
	"gophermart-loyalty/internal/gopherman/config/server"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/pkg/workerpool"
	"time"
)

type Client struct {
	db   *conn.DB
	Pool *workerpool.Pool
}

func NewClient(ctx context.Context, db *conn.DB, cfg *server.Config) *Client {
	reportPool := workerpool.NewPool(cfg.AccrualWorkerCount)
	reportPool.StartBg(ctx)
	return &Client{
		db:   db,
		Pool: reportPool,
	}
}
func (c *Client) StartPoolAccrual(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
