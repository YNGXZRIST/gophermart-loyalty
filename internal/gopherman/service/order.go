package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/internal/gopherman/repository"
	"gophermart-loyalty/pkg/luhn"
	"net/http"
	"strings"
)

// GetOrdersInput contains parameters for listing user orders.
type GetOrdersInput struct {
	UserID int64
}

// GetOrdersResponse contains service result and order list.
type GetOrdersResponse struct {
	Response
	Orders []*model.Order
}

// AddOrderInput contains parameters for adding an order.
type AddOrderInput struct {
	UserID  int64
	OrderID string
}

// GetOrders returns orders belonging to the requested user.
func (s *Service) GetOrders(ctx context.Context, in GetOrdersInput) GetOrdersResponse {
	orders, err := s.Rep.Order.GetByUserID(ctx, in.UserID)
	if err != nil {
		return GetOrdersResponse{
			Response: Response{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("get orders failed: %w", err),
			},
		}
	}
	return GetOrdersResponse{
		Response: Response{Code: http.StatusOK},
		Orders:   orders,
	}
}

// AddOrder validates and stores a new order for user.
func (s *Service) AddOrder(ctx context.Context, in AddOrderInput) Response {
	orderID := strings.TrimSpace(in.OrderID)
	if orderID == "" {
		return Response{
			Code: http.StatusBadRequest,
			Err:  errors.New("order id is required"),
		}
	}
	if !luhn.Validate(orderID) {
		return Response{
			Code: http.StatusUnprocessableEntity,
			Err:  errors.New("invalid order number"),
		}
	}
	err := s.Rep.Order.Add(ctx, in.UserID, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderExistsOwn) {
			return Response{Code: http.StatusOK}
		}
		if errors.Is(err, repository.ErrOrderExistsOther) {
			return Response{
				Code: http.StatusConflict,
				Err:  err,
			}
		}
		return Response{
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("add order failed: %w", err),
		}
	}
	return Response{Code: http.StatusAccepted}
}

// OrdersJSON encodes order list to JSON and normalizes nil slices.
func OrdersJSON(orders []*model.Order) ([]byte, error) {
	if orders == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(orders)
}
