package accrual

import (
	"context"
	"encoding/json"
	"gophermart-loyalty/internal/gopherman/constant"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/model"
	repo "gophermart-loyalty/internal/gopherman/repository"
	repoMock "gophermart-loyalty/internal/gopherman/repository/mock"
	"gophermart-loyalty/internal/gopherman/service"
	"gophermart-loyalty/pkg/httpretryable"
	"gophermart-loyalty/pkg/workerpool"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
)

func newMockConnDBForAccrual(t *testing.T) (*conn.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return &conn.DB{DB: db}, mock
}

func TestClient_doAccrualRequest_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want %s", r.Method, http.MethodGet)
		}
		w.Header().Set("Retry-After", "3")
		_, _ = w.Write([]byte(`{"order":"123","status":"processed","accrual":12.5}`))
	}))
	defer srv.Close()

	rc := httpretryable.NewRetryableClient()
	rc.RetryMax = 1
	c := &Client{
		httpClient: rc,
		accrualURL: srv.URL,
		logger:     zap.NewNop(),
	}

	res, err := c.doAccrualRequest(context.Background(), srv.URL+"/api/orders/123")
	if err != nil {
		t.Fatalf("doAccrualRequest error = %v", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, http.StatusOK)
	}
	if res.RetryAfter != "3" {
		t.Fatalf("RetryAfter = %q, want %q", res.RetryAfter, "3")
	}
	if string(res.Body) == "" {
		t.Fatalf("Body must be non-empty")
	}
}

func TestClient_handleSuccessResponse_DeletesInFlightAndParses(t *testing.T) {
	order := &model.Order{OrderID: "123", UserID: 10, Status: "NEW"}
	inFlight := map[string]*model.Order{"123": order}

	c := &Client{
		inFlight: inFlight,
	}

	body := []byte(`{"order":"123","status":"PROCESSED","accrual":12.5}`)
	got, err := c.handleSuccessResponse(body, order)
	if err != nil {
		t.Fatalf("handleSuccessResponse error = %v", err)
	}
	if got == nil {
		t.Fatalf("got is nil")
	}
	if _, ok := c.inFlight["123"]; ok {
		t.Fatalf("inFlight must delete orderID")
	}

	if got.Order != order {
		t.Fatalf("Order pointer mismatch")
	}
	if got.AccrualResponse == nil {
		t.Fatalf("AccrualResponse must not be nil")
	}
	if got.AccrualResponse.Status != constant.Processed {
		t.Fatalf("status = %q, want %q", got.AccrualResponse.Status, constant.Processed)
	}
	if got.AccrualResponse.Accrual == nil || *got.AccrualResponse.Accrual != 12.5 {
		t.Fatalf("accrual = %v, want 12.5", got.AccrualResponse.Accrual)
	}
	var parsed Response
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("json.Unmarshal sanity error: %v", err)
	}
}

func TestClient_updateOrder_registered_setsProcessing_and_incrementsBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrder := repoMock.NewMockOrderRepository(ctrl)
	mockUser := repoMock.NewMockUserRepository(ctrl)
	D, mockSQL := newMockConnDBForAccrual(t)
	mockSQL.ExpectBegin()
	mockSQL.ExpectCommit()

	order := &model.Order{
		ID:      1,
		UserID:  7,
		OrderID: "123",
		Status:  "NEW",
	}
	accrualVal := 55.5
	accrualResp := Response{
		Order:   "123",
		Status:  constant.Registered,
		Accrual: &accrualVal,
	}

	taskRes := workerpool.Task{
		Result: &TaskResult{
			Order:           order,
			AccrualResponse: &accrualResp,
		},
	}

	mockOrder.EXPECT().
		UpdateOrderAccrual(gomock.Any(), order).
		Return(nil)
	mockUser.EXPECT().
		IncrementBalance(gomock.Any(), order.UserID, accrualVal).
		Return(nil)

	c := &Client{
		ser:    service.NewService(D, repo.Repositories{User: mockUser, Order: mockOrder}),
		mu:     sync.Mutex{},
		logger: zap.NewNop(),
	}

	err := c.updateOrder(context.Background(), taskRes)
	if err != nil {
		t.Fatalf("updateOrder error = %v", err)
	}

	if order.Status != constant.Processing {
		t.Fatalf("order.Status = %q, want %q", order.Status, constant.Processing)
	}
	if order.Accrual == nil || *order.Accrual != accrualVal {
		t.Fatalf("order.Accrual = %v, want %v", order.Accrual, accrualVal)
	}

	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}
