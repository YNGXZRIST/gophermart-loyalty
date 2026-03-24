package accrual

import (
	"encoding/json"
	"gophermart-loyalty/internal/gopherman/errors/labelerrors"
	"gophermart-loyalty/internal/gopherman/model"
)

// Response describes an order status payload returned by accrual service.
type Response struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

func (c *Client) handleSuccessResponse(body []byte, order *model.Order) (*TaskResult, error) {
	var accResp Response
	if err := json.Unmarshal(body, &accResp); err != nil {
		c.removeFromInFlight(order.OrderID)
		return nil, labelerrors.NewLabelError(LabelAccrual+".Client.handleSuccess.Unmarshal", err)
	}
	c.removeFromInFlight(order.OrderID)
	return &TaskResult{
		Order:           order,
		AccrualResponse: &accResp,
	}, nil
}
