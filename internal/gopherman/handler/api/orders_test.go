package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"gophermart-loyalty/internal/gopherman/constant"
	"gophermart-loyalty/internal/gopherman/contextkey"
	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/internal/gopherman/repository"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

func TestHandler_AddOrder(t *testing.T) {
	t.Run("success_202_new_order", func(t *testing.T) {
		userID := int64(55)
		orderID := "79927398713"

		D, mockSQL := newMockConnDB(t)
		repos := repository.Repositories{
			User:       repository.NewUserRepository(D),
			Order:      repository.NewOrderRepository(D),
			Withdrawal: repository.NewWithdrawalRepository(D),
		}
		handler := NewHandler(D, repos, zap.NewNop())

		mockSQL.ExpectBegin()
		mockSQL.ExpectQuery(repository.OrderGetOwnerQuery).
			WithArgs(orderID).
			WillReturnError(sql.ErrNoRows)
		mockSQL.ExpectExec(repository.OrderAddOrderQuery).
			WithArgs(userID, orderID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mockSQL.ExpectCommit()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(orderID)))
		r = r.WithContext(contextkey.WithUserID(r.Context(), userID))

		handler.AddOrder(w, r)

		if got, want := w.Code, http.StatusAccepted; got != want {
			t.Fatalf("AddOrder status code = %d, want %d", got, want)
		}
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Fatalf("sqlmock expectations not met: %v", err)
		}
	})

	t.Run("success_200_already_own_order", func(t *testing.T) {
		userID := int64(55)
		orderID := "79927398713"

		D, mockSQL := newMockConnDB(t)
		repos := repository.Repositories{
			User:       repository.NewUserRepository(D),
			Order:      repository.NewOrderRepository(D),
			Withdrawal: repository.NewWithdrawalRepository(D),
		}
		handler := NewHandler(D, repos, zap.NewNop())

		mockSQL.ExpectBegin()
		mockSQL.ExpectQuery(repository.OrderGetOwnerQuery).
			WithArgs(orderID).
			WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(userID))
		mockSQL.ExpectRollback()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(orderID)))
		r = r.WithContext(contextkey.WithUserID(r.Context(), userID))

		handler.AddOrder(w, r)

		if got, want := w.Code, http.StatusOK; got != want {
			t.Fatalf("AddOrder status code = %d, want %d", got, want)
		}
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Fatalf("sqlmock expectations not met: %v", err)
		}
	})

	t.Run("conflict_409_other_user_order", func(t *testing.T) {
		userID := int64(55)
		orderID := "79927398713"
		otherUserID := int64(99)

		D, mockSQL := newMockConnDB(t)
		repos := repository.Repositories{
			User:       repository.NewUserRepository(D),
			Order:      repository.NewOrderRepository(D),
			Withdrawal: repository.NewWithdrawalRepository(D),
		}
		handler := NewHandler(D, repos, zap.NewNop())

		mockSQL.ExpectBegin()
		mockSQL.ExpectQuery(repository.OrderGetOwnerQuery).
			WithArgs(orderID).
			WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(otherUserID))
		mockSQL.ExpectRollback()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(orderID)))
		r = r.WithContext(contextkey.WithUserID(r.Context(), userID))

		handler.AddOrder(w, r)

		if got, want := w.Code, http.StatusConflict; got != want {
			t.Fatalf("AddOrder status code = %d, want %d", got, want)
		}
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Fatalf("sqlmock expectations not met: %v", err)
		}
	})

	t.Run("bad_request_400_empty_after_trim", func(t *testing.T) {
		userID := int64(55)

		D, mockSQL := newMockConnDB(t)
		repos := repository.Repositories{
			User:       repository.NewUserRepository(D),
			Order:      repository.NewOrderRepository(D),
			Withdrawal: repository.NewWithdrawalRepository(D),
		}
		handler := NewHandler(D, repos, zap.NewNop())

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("   ")))
		r = r.WithContext(contextkey.WithUserID(r.Context(), userID))

		handler.AddOrder(w, r)

		if got, want := w.Code, http.StatusBadRequest; got != want {
			t.Fatalf("AddOrder status code = %d, want %d", got, want)
		}
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Fatalf("sqlmock expectations not met: %v", err)
		}
	})

	t.Run("unprocessable_422_invalid_luhn", func(t *testing.T) {
		userID := int64(55)

		D, mockSQL := newMockConnDB(t)
		repos := repository.Repositories{
			User:       repository.NewUserRepository(D),
			Order:      repository.NewOrderRepository(D),
			Withdrawal: repository.NewWithdrawalRepository(D),
		}
		handler := NewHandler(D, repos, zap.NewNop())

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("12345")))
		r = r.WithContext(contextkey.WithUserID(r.Context(), userID))

		handler.AddOrder(w, r)

		if got, want := w.Code, http.StatusUnprocessableEntity; got != want {
			t.Fatalf("AddOrder status code = %d, want %d", got, want)
		}
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Fatalf("sqlmock expectations not met: %v", err)
		}
	})

	t.Run("unauthorized_401", func(t *testing.T) {
		D, mockSQL := newMockConnDB(t)
		repos := repository.Repositories{
			User:       repository.NewUserRepository(D),
			Order:      repository.NewOrderRepository(D),
			Withdrawal: repository.NewWithdrawalRepository(D),
		}
		handler := NewHandler(D, repos, zap.NewNop())

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("79927398713")))

		handler.AddOrder(w, r)

		if got, want := w.Code, http.StatusUnauthorized; got != want {
			t.Fatalf("AddOrder status code = %d, want %d", got, want)
		}
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Fatalf("sqlmock expectations not met: %v", err)
		}
	})

	t.Run("db_error_500_generic", func(t *testing.T) {
		userID := int64(55)
		orderID := "79927398713"

		D, mockSQL := newMockConnDB(t)
		repos := repository.Repositories{
			User:       repository.NewUserRepository(D),
			Order:      repository.NewOrderRepository(D),
			Withdrawal: repository.NewWithdrawalRepository(D),
		}
		handler := NewHandler(D, repos, zap.NewNop())

		mockSQL.ExpectBegin()
		mockSQL.ExpectQuery(repository.OrderGetOwnerQuery).
			WithArgs(orderID).
			WillReturnError(sql.ErrNoRows)
		mockSQL.ExpectExec(repository.OrderAddOrderQuery).
			WithArgs(userID, orderID).
			WillReturnError(errors.New("db error"))
		mockSQL.ExpectRollback()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(orderID)))
		r = r.WithContext(contextkey.WithUserID(r.Context(), userID))

		handler.AddOrder(w, r)

		if got, want := w.Code, http.StatusInternalServerError; got != want {
			t.Fatalf("AddOrder status code = %d, want %d", got, want)
		}
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Fatalf("sqlmock expectations not met: %v", err)
		}
	})
}

func TestHandler_GetOrders(t *testing.T) {
	t.Run("success_200_with_orders", func(t *testing.T) {
		userID := int64(55)
		createdAt1 := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
		createdAt2 := time.Date(2020, 1, 3, 4, 5, 6, 0, time.UTC)
		accrual := 12.34

		D, mockSQL := newMockConnDB(t)
		mockSQL.ExpectQuery(repository.OrderGetByUIDQuery).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "order_id", "status", "accrual", "created_at", "updated_at",
			}).
				AddRow(int64(1), "79927398713", "NEW", nil, createdAt1, time.Now()).
				AddRow(int64(2), "79927398714", "PROCESSED", accrual, createdAt2, time.Now()))

		repos := repository.Repositories{
			User:       repository.NewUserRepository(D),
			Order:      repository.NewOrderRepository(D),
			Withdrawal: repository.NewWithdrawalRepository(D),
		}
		handler := NewHandler(D, repos, zap.NewNop())

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(contextkey.WithUserID(r.Context(), userID))

		handler.GetOrders(w, r)

		if got, want := w.Code, http.StatusOK; got != want {
			t.Fatalf("GetOrders status code = %d, want %d", got, want)
		}
		if ct := w.Header().Get(constant.ContentTypeHeader); ct != constant.ApplicationJSON {
			t.Fatalf("Content-Type = %q, want %q", ct, constant.ApplicationJSON)
		}

		var resp []model.Order
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if len(resp) != 2 {
			t.Fatalf("orders len = %d, want 2; body=%s", len(resp), w.Body.String())
		}
		if resp[0].OrderID != "79927398713" || resp[0].Status != "NEW" {
			t.Fatalf("order[0] = %+v, want number=79927398713 status=NEW", resp[0])
		}
		if resp[0].Accrual != nil {
			t.Fatalf("order[0].Accrual must be nil, got %v", *resp[0].Accrual)
		}
		if resp[1].OrderID != "79927398714" || resp[1].Status != "PROCESSED" {
			t.Fatalf("order[1] = %+v, want number=79927398714 status=PROCESSED", resp[1])
		}
		if resp[1].Accrual == nil || *resp[1].Accrual != accrual {
			t.Fatalf("order[1].Accrual = %v, want %v", resp[1].Accrual, accrual)
		}

		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Fatalf("sqlmock expectations not met: %v", err)
		}
	})

	t.Run("empty_result_204", func(t *testing.T) {
		userID := int64(55)

		D, mockSQL := newMockConnDB(t)
		mockSQL.ExpectQuery(repository.OrderGetByUIDQuery).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "order_id", "status", "accrual", "created_at", "updated_at",
			}))

		repos := repository.Repositories{
			User:       repository.NewUserRepository(D),
			Order:      repository.NewOrderRepository(D),
			Withdrawal: repository.NewWithdrawalRepository(D),
		}
		handler := NewHandler(D, repos, zap.NewNop())

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(contextkey.WithUserID(r.Context(), userID))

		handler.GetOrders(w, r)

		if got, want := w.Code, http.StatusNoContent; got != want {
			t.Fatalf("GetOrders status code = %d, want %d", got, want)
		}
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Fatalf("sqlmock expectations not met: %v", err)
		}
	})

	t.Run("unauthorized_401", func(t *testing.T) {
		D, mockSQL := newMockConnDB(t)
		repos := repository.Repositories{
			User:       repository.NewUserRepository(D),
			Order:      repository.NewOrderRepository(D),
			Withdrawal: repository.NewWithdrawalRepository(D),
		}
		handler := NewHandler(D, repos, zap.NewNop())

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.GetOrders(w, r)

		if got, want := w.Code, http.StatusUnauthorized; got != want {
			t.Fatalf("GetOrders status code = %d, want %d", got, want)
		}
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Fatalf("sqlmock expectations not met: %v", err)
		}
	})

	t.Run("db_error_500", func(t *testing.T) {
		userID := int64(55)

		D, mockSQL := newMockConnDB(t)
		mockSQL.ExpectQuery(repository.OrderGetByUIDQuery).
			WithArgs(userID).
			WillReturnError(errors.New("db error"))

		repos := repository.Repositories{
			User:       repository.NewUserRepository(D),
			Order:      repository.NewOrderRepository(D),
			Withdrawal: repository.NewWithdrawalRepository(D),
		}
		handler := NewHandler(D, repos, zap.NewNop())

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(contextkey.WithUserID(r.Context(), userID))

		handler.GetOrders(w, r)

		if got, want := w.Code, http.StatusInternalServerError; got != want {
			t.Fatalf("GetOrders status code = %d, want %d", got, want)
		}
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Fatalf("sqlmock expectations not met: %v", err)
		}
	})
}
