// Package accrual polls the external accrual service for order statuses
// and applies accrual results to orders and user balances.
package accrual

import (
	"context"
	"fmt"
	"gophermart-loyalty/internal/gopherman/config/server"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/errors/labelerrors"
	"gophermart-loyalty/internal/gopherman/logger"
	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/internal/gopherman/repository"
	"gophermart-loyalty/internal/gopherman/service"
	"gophermart-loyalty/pkg/httpretryable"
	"gophermart-loyalty/pkg/workerpool"
	"io"
	"net/http"
	"sync"

	"go.uber.org/zap"
)

// Path is the relative endpoint for querying an order in the accrual service.
// Final request URL format: <accrual base URL> + Path + <order id>.
const Path = "/api/orders/"

// Registered means the order is accepted by the accrual service.
//
// The service may later transition it to Processing or a final state.
const (
	Registered = "REGISTERED"
	// Processing means accrual calculation is in progress.
	Processing = "PROCESSING"
	// Processed means accrual calculation is completed successfully.
	Processed = "PROCESSED"
	// Invalid means the order was rejected by the accrual service.
	Invalid = "INVALID"
)

// LabelAccrual is a common error/trace label for accrual package operations.
const (
	LabelAccrual = "ACCRUAL"
	// LoggerType identifies the logger configuration for accrual workers.
	LoggerType = "accrual"
)

// Client fetches order accrual statuses from the external service,
// deduplicates in-flight requests, and applies results via service layer.
type Client struct {
	ser        *service.Service
	Pool       *workerpool.Pool
	httpClient *httpretryable.RetryableClient
	inFlight   map[string]*model.Order
	mu         sync.Mutex
	accrualURL string
	logger     *zap.Logger
}
type accrualResult struct {
	StatusCode int
	Body       []byte
}

// NewClient creates an accrual client with background worker pool,
// retryable HTTP transport, and package-scoped logger.
func NewClient(ctx context.Context, db *conn.DB, repos repository.Repositories, cfg *server.Config) (*Client, error) {
	newService := service.NewService(db, repos)
	reportPool := workerpool.NewPool(cfg.AccrualWorkerCount)
	httpClient := httpretryable.NewRetryableClient()
	httpClient.RetryMax = 10
	reportPool.StartBg(ctx)
	lgr, err := logger.Initialize(cfg.Mode, LoggerType)
	if err != nil {
		return nil, labelerrors.NewLabelError(LabelAccrual+".NewClient.Logger", err)
	}
	return &Client{
		ser:        newService,
		Pool:       reportPool,
		httpClient: httpClient,
		accrualURL: cfg.AccrualAddress,
		logger:     lgr,
		inFlight:   make(map[string]*model.Order, 1000),
	}, nil
}

func (c *Client) sendRequestToAccrual(ctx context.Context, order *model.Order) (*TaskResult, error) {
	url := c.accrualURL + Path + order.OrderID
	select {
	case <-ctx.Done():
		c.removeFromInFlight(order.OrderID)
		return nil, labelerrors.NewLabelError(LabelAccrual+".Client.SendRequest.Context", ctx.Err())
	default:
	}
	result, err := c.doAccrualRequest(ctx, url)
	if err != nil {
		c.removeFromInFlight(order.OrderID)
		return nil, err
	}
	if result.StatusCode == http.StatusOK {
		return c.handleSuccessResponse(result.Body, order)
	}
	c.removeFromInFlight(order.OrderID)
	return nil, labelerrors.NewLabelError(LabelAccrual+".Client.SendRequest.BadStatus", fmt.Errorf("accrual HTTP status %d", result.StatusCode))
}

func (c *Client) removeFromInFlight(orderID string) {
	c.mu.Lock()
	delete(c.inFlight, orderID)
	c.mu.Unlock()
}

func (c *Client) doAccrualRequest(ctx context.Context, url string) (*accrualResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, labelerrors.NewLabelError(LabelAccrual+".Client.doAccrualRequest.NewRequest", err)
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, labelerrors.NewLabelError(LabelAccrual+".Client.doAccrualRequest.Do", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, labelerrors.NewLabelError(LabelAccrual+".Client.doAccrualRequest.ReadBody", err)
	}
	return &accrualResult{
		StatusCode: resp.StatusCode,
		Body:       body,
	}, nil
}
