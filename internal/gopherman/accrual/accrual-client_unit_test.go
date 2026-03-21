package accrual

import (
	"encoding/json"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/model"
	repo "gophermart-loyalty/internal/gopherman/repository"
	"gophermart-loyalty/internal/gopherman/service"
	"gophermart-loyalty/pkg/httpretryable"
	"gophermart-loyalty/pkg/workerpool"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
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

	res, err := c.doAccrualRequest(t.Context(), srv.URL+"/api/orders/123")
	if err != nil {
		t.Fatalf("doAccrualRequest error = %v", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, http.StatusOK)
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
	if got.AccrualResponse.Status != Processed {
		t.Fatalf("status = %q, want %q", got.AccrualResponse.Status, Processed)
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
	D, mockSQL := newMockConnDBForAccrual(t)
	orderRepo := repo.NewOrderRepository(D)
	userRepo := repo.NewUserRepository(D)

	order := &model.Order{
		ID:      1,
		UserID:  7,
		OrderID: "123",
		Status:  "NEW",
	}
	accrualVal := 55.5
	accrualResp := Response{
		Order:   "123",
		Status:  Registered,
		Accrual: &accrualVal,
	}

	taskRes := workerpool.Task{
		Result: &TaskResult{
			Order:           order,
			AccrualResponse: &accrualResp,
		},
	}

	mockSQL.ExpectBegin()
	mockSQL.ExpectExec(repo.OrderUpdatePendingOrderQuery).
		WithArgs(Processing, accrualVal, order.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mockSQL.ExpectQuery(repo.UserGetByIDQuery).
		WithArgs(order.UserID).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn"}).
				AddRow(order.UserID, "test", "pass", time.Now().Add(-time.Hour), time.Now(), "old-ip", 100.0, 20.0),
		)
	mockSQL.ExpectExec(repo.UserIncrementBalanceQuery).
		WithArgs(155.5, order.UserID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mockSQL.ExpectCommit()

	c := &Client{
		ser:    service.NewService(D, repo.Repositories{User: userRepo, Order: orderRepo}),
		mu:     sync.Mutex{},
		logger: zap.NewNop(),
	}

	err := c.updateOrder(t.Context(), taskRes)
	if err != nil {
		t.Fatalf("updateOrder error = %v", err)
	}

	if order.Status != Processing {
		t.Fatalf("order.Status = %q, want %q", order.Status, Processing)
	}
	if order.Accrual == nil || *order.Accrual != accrualVal {
		t.Fatalf("order.Accrual = %v, want %v", order.Accrual, accrualVal)
	}

	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}
