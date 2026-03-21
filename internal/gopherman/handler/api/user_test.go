package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"gophermart-loyalty/internal/gopherman/auth/password"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/internal/gopherman/repository"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

func TestHandler_Login(t *testing.T) {
	userID := int64(10)
	login := "test"
	ip := "127.0.0.1"

	hash, err := password.Hash("correct-pass")
	if err != nil {
		t.Fatalf("password.Hash() error: %v", err)
	}

	db, m, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	D := &conn.DB{DB: db}
	m.ExpectQuery(regexp.QuoteMeta(repository.UserGetByLoginQuery)).
		WithArgs(login).
		WillReturnRows(sqlmock.NewRows([]string{"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn"}).
			AddRow(userID, login, hash, time.Now(), time.Now(), ip, 0.0, 0.0))
	m.ExpectQuery(regexp.QuoteMeta(repository.UserGetByIDQuery)).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn"}).
			AddRow(userID, login, hash, time.Now(), time.Now(), ip, 0.0, 0.0))
	m.ExpectExec(regexp.QuoteMeta(repository.UserUpdateLastIPQuery)).
		WithArgs(ip, userID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	m.ExpectExec(regexp.QuoteMeta(repository.UserUpsertSessionQuery)).
		WithArgs(sqlmock.AnyArg(), userID, sqlmock.AnyArg(), ip).
		WillReturnResult(sqlmock.NewResult(1, 1))

	handler := NewHandler(D, repository.Repositories{User: repository.NewUserRepository(D)}, zap.NewNop())

	reqBody, _ := json.Marshal(model.RegisterRequest{Login: login, Pass: "correct-pass"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
	req.RemoteAddr = ip + ":1234"

	w := httptest.NewRecorder()
	handler.Login(w, req.WithContext(t.Context()))

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("Login status code = %d, want %d", got, want)
	}
	if got := w.Header().Get("Authorization"); got == "" {
		t.Fatal("Authorization header is empty")
	}
	if err := m.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}

func TestHandler_Register(t *testing.T) {
	userID := int64(7)
	login := "test2"
	ip := "127.0.0.1"

	db, m, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	D := &conn.DB{DB: db}
	m.ExpectQuery(regexp.QuoteMeta(repository.UserGetByLoginQuery)).
		WithArgs(login).
		WillReturnError(sql.ErrNoRows)
	m.ExpectQuery(regexp.QuoteMeta(repository.UserRegisterQuery)).
		WithArgs(login, sqlmock.AnyArg(), ip).
		WillReturnRows(sqlmock.NewRows([]string{"id", "login", "pass", "created_at", "updated_at", "last_login_ip"}).
			AddRow(userID, login, "hash", time.Now(), time.Now(), ip))
	m.ExpectQuery(regexp.QuoteMeta(repository.UserGetByIDQuery)).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn"}).
			AddRow(userID, login, "hash", time.Now(), time.Now(), ip, 0.0, 0.0))
	m.ExpectExec(regexp.QuoteMeta(repository.UserUpdateLastIPQuery)).
		WithArgs(ip, userID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	m.ExpectExec(regexp.QuoteMeta(repository.UserUpsertSessionQuery)).
		WithArgs(sqlmock.AnyArg(), userID, sqlmock.AnyArg(), ip).
		WillReturnResult(sqlmock.NewResult(1, 1))
	handler := NewHandler(D, repository.Repositories{User: repository.NewUserRepository(D)}, zap.NewNop())

	reqBody, _ := json.Marshal(model.RegisterRequest{Login: login, Pass: "secret-pass"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
	req.RemoteAddr = ip + ":1234"

	w := httptest.NewRecorder()
	handler.Register(w, req.WithContext(t.Context()))

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("Register status code = %d, want %d", got, want)
	}
	if got := w.Header().Get("Authorization"); got == "" {
		t.Fatal("Authorization header is empty")
	}

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("not-json")))
	req2.RemoteAddr = ip + ":1234"
	handler.Register(w2, req2.WithContext(t.Context()))
	if got, want := w2.Code, http.StatusBadRequest; got != want {
		t.Fatalf("Register(bad json) status code = %d, want %d", got, want)
	}
	if err := m.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}
