package accrual

import (
	"context"
	"fmt"
	"gophermart-loyalty/internal/gopherman/errors/labelerrors"
	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/pkg/workerpool"
	"time"

	"go.uber.org/zap"
)

// TaskResult carries parsed accrual response and source order for processing.
type TaskResult struct {
	Order           *model.Order
	AccrualResponse *Response
	Error           error
}

// StartPoolAccrual periodically fetches pending orders
// and schedules accrual polling tasks for unique in-flight orders.
func (c *Client) StartPoolAccrual(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.mu.Lock()
			c.mu.Unlock()
			list, err := c.ser.Rep.Order.GetOrdersPendingAccrual(ctx)
			if err != nil {
				c.logger.Error("Error getting orders", zap.Error(labelerrors.NewLabelError(LabelAccrual+".Client.PendingOrders", err)))
			}
			for _, o := range list {
				c.mu.Lock()
				if _, ok := c.inFlight[o.OrderID]; !ok {
					c.inFlight[o.OrderID] = o
					c.mu.Unlock()
					task := workerpool.NewTask(func(x any) (any, error) {
						order := x.(*model.Order)
						accrual, err := c.sendRequestToAccrual(ctx, order)
						return accrual, err
					})
					task.NeedResult = true
					task.Result = o
					c.Pool.Add(task)
				} else {
					c.mu.Unlock()
				}
			}
		}
	}
}

// CollectResults consumes completed worker tasks
// and applies accrual updates to the related orders.
func (c *Client) CollectResults(ctx context.Context) {
	resultCh := make(chan workerpool.Task, 1)
	go func() {
		for {
			res := c.Pool.Get()
			select {
			case resultCh <- res:
			case <-ctx.Done():
				return
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case res := <-resultCh:
			if res.Err != nil {
				c.logger.Error("collect result income error", zap.Error(labelerrors.NewLabelError(LabelAccrual+".Client.Task", res.Err)))
				continue
			}
			err := c.updateOrder(ctx, res)
			if err != nil {
				c.logger.Error("error updating order", zap.Error(labelerrors.NewLabelError(LabelAccrual+".Client.UpdateOrder", err)))
			}
		}
	}
}
func (c *Client) updateOrder(ctx context.Context, taskRes workerpool.Task) error {
	res, ok := taskRes.Result.(*TaskResult)
	if !ok {
		return fmt.Errorf("type assertion failed")
	}
	accrualResponse := res.AccrualResponse
	order := res.Order
	status := accrualResponse.Status
	if status == Registered {
		status = Processing
	}
	order.Status = status
	order.Accrual = accrualResponse.Accrual
	if err := c.ser.ApplyAccrualResult(ctx, order); err != nil {
		return labelerrors.NewLabelError(LabelAccrual+".Client.updateOrder.Apply", err)
	}
	return nil
}
