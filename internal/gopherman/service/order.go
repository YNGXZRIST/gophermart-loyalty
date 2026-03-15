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

type GetOrdersInput struct {
	UserID int64
}

type GetOrdersResponse struct {
	Response
	Orders []*model.Order
}

type AddOrderInput struct {
	UserID  int64
	OrderID string
}

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

func OrdersJSON(orders []*model.Order) ([]byte, error) {
	if orders == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(orders)
}
