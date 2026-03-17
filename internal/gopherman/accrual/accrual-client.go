package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"gophermart-loyalty/internal/gopherman/config/server"
	"gophermart-loyalty/internal/gopherman/constant"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/errors/labelerrors"
	"gophermart-loyalty/internal/gopherman/logger"
	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/internal/gopherman/repository"
	"gophermart-loyalty/internal/gopherman/service"
	"gophermart-loyalty/pkg/workerpool"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Client struct {
	db         *conn.DB
	ser        *service.Service
	Pool       *workerpool.Pool
	httpClient *http.Client
	inFlight   map[string]*model.Order
	mu         sync.Mutex
	accrualURL string
	RetryAfter time.Time
	logger     *zap.Logger
}
type TaskResult struct {
	Order           *model.Order
	AccrualResponse *Response
	Error           error
}
type Response struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}
type accrualResult struct {
	StatusCode int
	Body       []byte
	RetryAfter string
}

func NewClient(ctx context.Context, db *conn.DB, repos repository.Repositories, cfg *server.Config) (*Client, error) {
	newService := service.NewService(db, repos)
	reportPool := workerpool.NewPool(cfg.AccrualWorkerCount)
	httpClient := &http.Client{}
	reportPool.StartBg(ctx)
	lgr, err := logger.Initialize(cfg.Mode, constant.AccrualType)
	if err != nil {
		return nil, labelerrors.NewLabelError(constant.LabelAccrual+".NewClient.Logger", err)
	}
	return &Client{
		db:         db,
		ser:        newService,
		Pool:       reportPool,
		httpClient: httpClient,
		accrualURL: cfg.AccrualAddress,
		logger:     lgr,
		inFlight:   make(map[string]*model.Order, 1000),
	}, nil
}

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
				c.logger.Error("Error getting orders", zap.Error(labelerrors.NewLabelError(constant.LabelAccrual+".Client.PendingOrders", err)))
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
				c.logger.Error("collect result income error", zap.Error(labelerrors.NewLabelError(constant.LabelAccrual+".Client.Task", res.Err)))
				continue
			}
			err := c.updateOrder(ctx, res)
			if err != nil {
				c.logger.Error("error updating order", zap.Error(labelerrors.NewLabelError(constant.LabelAccrual+".Client.UpdateOrder", err)))
			}
		}
	}
}
func (c *Client) sendRequestToAccrual(ctx context.Context, order *model.Order) (*TaskResult, error) {
	url := c.accrualURL + "/api/orders/" + order.OrderID
	for {
		select {
		case <-ctx.Done():
			return nil, labelerrors.NewLabelError(constant.LabelAccrual+".Client.SendRequest.Context", ctx.Err())
		default:
		}
		c.mu.Lock()
		until := c.RetryAfter
		c.mu.Unlock()
		if until.After(time.Now()) {
			timer := time.NewTimer(time.Until(until))
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, labelerrors.NewLabelError(constant.LabelAccrual+".Client.SendRequest.Context", ctx.Err())
			case <-timer.C:
			}
		}
		result, err := c.doAccrualRequest(ctx, url)
		if err != nil {
			return nil, err
		}
		switch result.StatusCode {
		case http.StatusOK:
			return c.handleSuccessResponse(result.Body, order)
		case http.StatusTooManyRequests:
			sec := c.setRetryTimeout(result.RetryAfter)
			err := waitOrCancel(ctx, time.Duration(sec)*time.Second)
			if err != nil {
				return nil, labelerrors.NewLabelError(constant.LabelAccrual+".Client.SendRequest.WaitRetryAfter", err)
			}
			continue
		default:
			err := waitOrCancel(ctx, 2*time.Second)
			if err != nil {
				return nil, labelerrors.NewLabelError(constant.LabelAccrual+".Client.SendRequest.WaitDefault", err)
			}
			continue
		}

	}
}

func (c *Client) doAccrualRequest(ctx context.Context, url string) (*accrualResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, labelerrors.NewLabelError(constant.LabelAccrual+".Client.doAccrualRequest.NewRequest", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, labelerrors.NewLabelError(constant.LabelAccrual+".Client.doAccrualRequest.Do", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, labelerrors.NewLabelError(constant.LabelAccrual+".Client.doAccrualRequest.ReadBody", err)
	}
	return &accrualResult{
		StatusCode: resp.StatusCode,
		Body:       body,
		RetryAfter: resp.Header.Get("Retry-After"),
	}, nil
}

func (c *Client) handleSuccessResponse(body []byte, order *model.Order) (*TaskResult, error) {
	var accResp Response
	if err := json.Unmarshal(body, &accResp); err != nil {
		return nil, labelerrors.NewLabelError(constant.LabelAccrual+".Client.handleSuccess.Unmarshal", err)
	}
	c.mu.Lock()
	delete(c.inFlight, order.OrderID)
	c.mu.Unlock()
	return &TaskResult{
		Order:           order,
		AccrualResponse: &accResp,
	}, nil
}
func (c *Client) setRetryTimeout(retryAfter string) int {
	sec := 1
	if n, err := strconv.Atoi(retryAfter); err == nil && n > 0 {
		sec = n
	}
	fmt.Println("retrying after", sec)
	c.mu.Lock()
	c.RetryAfter = time.Now().Add(time.Duration(sec) * time.Second)
	c.mu.Unlock()
	return sec
}
func waitOrCancel(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
func (c *Client) updateOrder(ctx context.Context, taskRes workerpool.Task) error {
	res := taskRes.Result.(*TaskResult)
	accrualResponse := res.AccrualResponse
	order := res.Order
	status := accrualResponse.Status
	if status == constant.Registered {
		status = constant.Processing
	}
	order.Status = status
	order.Accrual = accrualResponse.Accrual
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return labelerrors.NewLabelError(constant.LabelAccrual+".Client.updateOrder.BeginTx", fmt.Errorf("cannot begin transaction %w", err))
	}
	defer tx.Rollback()
	err = c.ser.Rep.Order.UpdateOrderAccrual(ctx, tx, order)
	if err != nil {
		return labelerrors.NewLabelError(constant.LabelAccrual+".Client.updateOrder.UpdateAccrual", fmt.Errorf("cannot update order accrual %w", err))

	}
	if order.Accrual != nil {
		err = c.ser.Rep.User.IncrementBalance(ctx, tx, order.UserID, *order.Accrual)
		if err != nil {
			return labelerrors.NewLabelError(constant.LabelAccrual+".Client.updateOrder.IncrementBalance", fmt.Errorf("cannot increment balance for user %w", err))
		}
	}
	if err := tx.Commit(); err != nil {
		return labelerrors.NewLabelError(constant.LabelAccrual+".Client.updateOrder.Commit", err)
	}
	return nil
}
