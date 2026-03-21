package api

import (
	"bytes"
	"encoding/json"
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

func TestHandler_GetBalance(t *testing.T) {
	user := &model.User{
		ID: 4,
	}
	D, mockSQL := newMockConnDB(t)
	mockSQL.ExpectQuery(
		repository.UserGetByIDQuery,
	).
		WithArgs(user.ID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn",
		}).AddRow(
			user.ID,
			"test",
			"hashed-pass",
			time.Now().Add(-time.Hour),
			time.Now(),
			"1.2.3.4",
			100.5,
			10.5,
		))
	repos := repository.Repositories{
		User:       repository.NewUserRepository(D),
		Order:      repository.NewOrderRepository(D),
		Withdrawal: repository.NewWithdrawalRepository(D),
	}
	handler := NewHandler(D, repos, zap.NewNop())
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(contextkey.WithUserID(r.Context(), user.ID))
	handler.GetBalance(w, r)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("GetBalance status code = %d, want %d", got, want)
	}
	if ct := w.Header().Get(constant.ContentTypeHeader); ct != constant.ApplicationJSON {
		t.Fatalf("Content-Type = %q, want %q", ct, constant.ApplicationJSON)
	}

	var resp model.BalanceResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response json unmarshal error: %v", err)
	}
	if resp.Current != 100.5 || resp.Withdrawn != 10.5 {
		t.Fatalf("balance response = %+v, want current=100.5 withdrawn=10.5", resp)
	}

	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}

func TestHandler_GetBalance_Unauthorized(t *testing.T) {
	D, _ := newMockConnDB(t)
	repos := repository.Repositories{
		User:       repository.NewUserRepository(D),
		Order:      repository.NewOrderRepository(D),
		Withdrawal: repository.NewWithdrawalRepository(D),
	}
	handler := NewHandler(D, repos, zap.NewNop())

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.GetBalance(w, r)

	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Fatalf("GetBalance status code = %d, want %d", got, want)
	}
}

func TestHandler_MakeWithdraw(t *testing.T) {
	t.Parallel()
	t.Run("success_200", func(t *testing.T) {
		userID := int64(4)
		orderID := "5062821234567892"
		amount := 30.0

		D, mockSQL := newMockConnDB(t)
		repos := repository.Repositories{
			User:       repository.NewUserRepository(D),
			Order:      repository.NewOrderRepository(D),
			Withdrawal: repository.NewWithdrawalRepository(D),
		}

		handler := NewHandler(D, repos, zap.NewNop())

		createdAt := time.Now().Add(-time.Hour)
		updatedAt := time.Now()
		lastIP := "1.2.3.4"
		initialBalance := 100.0
		initialWithdrawn := 20.0

		mockSQL.ExpectQuery(
			repository.UserGetByIDQuery,
		).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn",
			}).AddRow(
				userID,
				"test",
				"hashed-pass",
				createdAt,
				updatedAt,
				lastIP,
				initialBalance,
				initialWithdrawn,
			))

		mockSQL.ExpectBegin()

		mockSQL.ExpectExec(
			repository.WithdrawalAddQuery,
		).
			WithArgs(userID, orderID, amount).
			WillReturnResult(sqlmock.NewResult(1, 1))

		updatedBalance := initialBalance - amount
		updatedWithdrawn := initialWithdrawn + amount
		mockSQL.ExpectExec(
			repository.UserIncrementWithdrawnQuery,
		).
			WithArgs(updatedBalance, updatedWithdrawn, userID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mockSQL.ExpectCommit()

		reqBody, err := json.Marshal(WithdrawalRequest{OrderID: orderID, Amount: amount})
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
		r = r.WithContext(contextkey.WithUserID(r.Context(), userID))

		handler.MakeWithdraw(w, r)

		if got, want := w.Code, http.StatusOK; got != want {
			t.Fatalf("MakeWithdraw status code = %d, want %d", got, want)
		}

		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Fatalf("sqlmock expectations not met: %v", err)
		}
	})

	t.Run("unauthorized_401", func(t *testing.T) {
		D, _ := newMockConnDB(t)
		repos := repository.Repositories{
			User:       repository.NewUserRepository(D),
			Order:      repository.NewOrderRepository(D),
			Withdrawal: repository.NewWithdrawalRepository(D),
		}
		handler := NewHandler(D, repos, zap.NewNop())

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"order":"1","sum":1}`)))
		handler.MakeWithdraw(w, r)

		if got, want := w.Code, http.StatusUnauthorized; got != want {
			t.Fatalf("MakeWithdraw status code = %d, want %d", got, want)
		}
	})

	t.Run("bad_json_400", func(t *testing.T) {
		D, mockSQL := newMockConnDB(t)
		repos := repository.Repositories{
			User:       repository.NewUserRepository(D),
			Order:      repository.NewOrderRepository(D),
			Withdrawal: repository.NewWithdrawalRepository(D),
		}
		handler := NewHandler(D, repos, zap.NewNop())

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`not-json`)))
		r = r.WithContext(contextkey.WithUserID(r.Context(), 1))

		handler.MakeWithdraw(w, r)

		if got, want := w.Code, http.StatusBadRequest; got != want {
			t.Fatalf("MakeWithdraw status code = %d, want %d", got, want)
		}

		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Fatalf("sqlmock expectations not met: %v", err)
		}
	})
}
