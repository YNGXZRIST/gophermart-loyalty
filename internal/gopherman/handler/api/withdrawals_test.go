package api

import (
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

func TestHandler_GetWithdrawals(t *testing.T) {
	t.Run("success_200", func(t *testing.T) {
		userID := int64(7)
		createdAt := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
		updatedAt := time.Date(2020, 1, 3, 4, 5, 6, 0, time.UTC)

		D, mockSQL := newMockConnDB(t)
		mockSQL.ExpectQuery(
			repository.WithdrawalGetByUserIDQuery,
		).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"order_id", "sum", "created_at", "updated_at"}).
				AddRow("w1", 100.0, createdAt, updatedAt))

		repos := repository.Repositories{
			User:       repository.NewUserRepository(D),
			Order:      repository.NewOrderRepository(D),
			Withdrawal: repository.NewWithdrawalRepository(D),
		}
		handler := NewHandler(D, repos, zap.NewNop())

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(contextkey.WithUserID(r.Context(), userID))

		handler.GetWithdrawals(w, r)

		if got, want := w.Code, http.StatusOK; got != want {
			t.Fatalf("GetWithdrawals status code = %d, want %d", got, want)
		}
		if ct := w.Header().Get(constant.ContentTypeHeader); ct != constant.ApplicationJSON {
			t.Fatalf("Content-Type = %q, want %q", ct, constant.ApplicationJSON)
		}

		var resp []model.Withdrawal
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if len(resp) != 1 || resp[0].OrderID != "w1" || resp[0].Sum != 100.0 || !resp[0].UpdatedAt.Equal(updatedAt) {
			t.Fatalf("response = %+v, want orderID=w1 sum=100 processed_at=%v", resp, updatedAt)
		}

		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Fatalf("sqlmock expectations not met: %v", err)
		}
	})

	t.Run("empty_result_204", func(t *testing.T) {
		userID := int64(7)
		D, mockSQL := newMockConnDB(t)

		mockSQL.ExpectQuery(
			repository.WithdrawalGetByUserIDQuery,
		).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"order_id", "sum", "created_at", "updated_at"}))

		repos := repository.Repositories{
			User:       repository.NewUserRepository(D),
			Order:      repository.NewOrderRepository(D),
			Withdrawal: repository.NewWithdrawalRepository(D),
		}
		handler := NewHandler(D, repos, zap.NewNop())

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(contextkey.WithUserID(r.Context(), userID))

		handler.GetWithdrawals(w, r)

		if got, want := w.Code, http.StatusNoContent; got != want {
			t.Fatalf("GetWithdrawals status code = %d, want %d", got, want)
		}
	})

	t.Run("unauthorized_401", func(t *testing.T) {
		D, mockSQL := newMockConnDB(t)
		_ = D

		repos := repository.Repositories{
			User:       repository.NewUserRepository(D),
			Order:      repository.NewOrderRepository(D),
			Withdrawal: repository.NewWithdrawalRepository(D),
		}
		handler := NewHandler(D, repos, zap.NewNop())

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		handler.GetWithdrawals(w, r)
		if got, want := w.Code, http.StatusUnauthorized; got != want {
			t.Fatalf("GetWithdrawals status code = %d, want %d", got, want)
		}

		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Fatalf("sqlmock expectations not met: %v", err)
		}
	})

	t.Run("db_error_500", func(t *testing.T) {
		userID := int64(7)
		D, mockSQL := newMockConnDB(t)

		mockSQL.ExpectQuery(
			repository.WithdrawalGetByUserIDQuery,
		).
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

		handler.GetWithdrawals(w, r)
		if got, want := w.Code, http.StatusInternalServerError; got != want {
			t.Fatalf("GetWithdrawals status code = %d, want %d", got, want)
		}

		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Fatalf("sqlmock expectations not met: %v", err)
		}
	})
}
