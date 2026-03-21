package service

import (
	"database/sql"
	"errors"
	"gophermart-loyalty/internal/gopherman/auth/password"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/internal/gopherman/repository"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestService_Register(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	req := model.RegisterRequest{Login: "user1", Pass: "secret12"}
	ip := "127.0.0.1"

	tests := []struct {
		name       string
		setup      func(m sqlmock.Sqlmock)
		wantCode   int
		wantErr    bool
		wantToken  bool
		tokenValue string
	}{
		{
			name: "GetByLogin database error",
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(repository.UserGetByLoginQuery)).
					WithArgs(req.Login).
					WillReturnError(errors.New("db down"))
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  true,
		},
		{
			name: "login already taken",
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(repository.UserGetByLoginQuery)).
					WithArgs(req.Login).
					WillReturnRows(sqlmock.NewRows([]string{"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn"}).
						AddRow(int64(1), req.Login, "hash", time.Now(), time.Now(), "127.0.0.1", 0.0, 0.0))
			},
			wantCode: http.StatusConflict,
			wantErr:  true,
		},
		{
			name: "new user ErrNoRows then register fails",
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(repository.UserGetByLoginQuery)).
					WithArgs(req.Login).
					WillReturnError(sql.ErrNoRows)
				m.ExpectQuery(regexp.QuoteMeta(repository.UserRegisterQuery)).
					WithArgs(req.Login, sqlmock.AnyArg(), ip).
					WillReturnError(errors.New("insert failed"))
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  true,
		},
		{
			name: "session creation fails",
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(repository.UserGetByLoginQuery)).
					WithArgs(req.Login).
					WillReturnError(sql.ErrNoRows)
				m.ExpectQuery(regexp.QuoteMeta(repository.UserRegisterQuery)).
					WithArgs(req.Login, sqlmock.AnyArg(), ip).
					WillReturnRows(sqlmock.NewRows([]string{"id", "login", "pass", "created_at", "updated_at", "last_login_ip"}).
						AddRow(int64(42), req.Login, "hash", time.Now(), time.Now(), ip))
				m.ExpectQuery(regexp.QuoteMeta(repository.UserGetByIDQuery)).
					WithArgs(int64(42)).
					WillReturnRows(sqlmock.NewRows([]string{"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn"}).
						AddRow(int64(42), req.Login, "hash", time.Now(), time.Now(), ip, 0.0, 0.0))
				m.ExpectExec(regexp.QuoteMeta(repository.UserUpdateLastIPQuery)).
					WithArgs(ip, int64(42)).
					WillReturnError(errors.New("session error"))
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  true,
		},
		{
			name: "success",
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(repository.UserGetByLoginQuery)).
					WithArgs(req.Login).
					WillReturnError(sql.ErrNoRows)
				m.ExpectQuery(regexp.QuoteMeta(repository.UserRegisterQuery)).
					WithArgs(req.Login, sqlmock.AnyArg(), ip).
					WillReturnRows(sqlmock.NewRows([]string{"id", "login", "pass", "created_at", "updated_at", "last_login_ip"}).
						AddRow(int64(7), req.Login, "hash", time.Now(), time.Now(), ip))
				m.ExpectQuery(regexp.QuoteMeta(repository.UserGetByIDQuery)).
					WithArgs(int64(7)).
					WillReturnRows(sqlmock.NewRows([]string{"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn"}).
						AddRow(int64(7), req.Login, "hash", time.Now(), time.Now(), ip, 0.0, 0.0))
				m.ExpectExec(regexp.QuoteMeta(repository.UserUpdateLastIPQuery)).
					WithArgs(ip, int64(7)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec(regexp.QuoteMeta(repository.UserUpsertSessionQuery)).
					WithArgs(sqlmock.AnyArg(), int64(7), sqlmock.AnyArg(), ip).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantCode:  http.StatusOK,
			wantToken: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db, m, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock.New: %v", err)
			}
			defer db.Close()
			D := &conn.DB{DB: db}
			tt.setup(m)
			s := NewService(D, repository.Repositories{User: repository.NewUserRepository(D), Order: repository.NewOrderRepository(D), Withdrawal: repository.NewWithdrawalRepository(D)})
			out := s.Register(ctx, RegisterInput{Req: req, IP: ip})
			if out.Code != tt.wantCode {
				t.Errorf("Code = %d, want %d", out.Code, tt.wantCode)
			}
			if tt.wantErr && out.Err == nil {
				t.Error("expected error")
			}
			if !tt.wantErr && out.Err != nil {
				t.Errorf("unexpected err: %v", out.Err)
			}
			if tt.wantToken && out.Token == "" {
				t.Error("expected non-empty token")
			}
			if err := m.ExpectationsWereMet(); err != nil {
				t.Fatalf("sqlmock expectations not met: %v", err)
			}
		})
	}
}

func TestService_Login(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	req := model.RegisterRequest{Login: "test", Pass: "correct"}
	ip := "10.0.0.1"
	hash, err := password.Hash(req.Pass)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		setup      func(m sqlmock.Sqlmock)
		wantCode   int
		wantErr    bool
		wantToken  bool
		tokenValue string
	}{
		{
			name: "GetByLogin error",
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(repository.UserGetByLoginQuery)).
					WithArgs(req.Login).
					WillReturnError(errors.New("timeout"))
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  true,
		},
		{
			name: "user not found",
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(repository.UserGetByLoginQuery)).
					WithArgs(req.Login).
					WillReturnError(sql.ErrNoRows)
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  true,
		},
		{
			name: "wrong password",
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(repository.UserGetByLoginQuery)).
					WithArgs(req.Login).
					WillReturnRows(sqlmock.NewRows([]string{"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn"}).
						AddRow(int64(1), req.Login, hash, time.Now(), time.Now(), "127.0.0.1", 0.0, 0.0))
			},
			wantCode: http.StatusUnauthorized,
			wantErr:  true,
		},
		{
			name: "CreateSession fails",
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(repository.UserGetByLoginQuery)).
					WithArgs(req.Login).
					WillReturnRows(sqlmock.NewRows([]string{"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn"}).
						AddRow(int64(2), req.Login, hash, time.Now(), time.Now(), "127.0.0.1", 0.0, 0.0))
				m.ExpectQuery(regexp.QuoteMeta(repository.UserGetByIDQuery)).
					WithArgs(int64(2)).
					WillReturnRows(sqlmock.NewRows([]string{"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn"}).
						AddRow(int64(2), req.Login, hash, time.Now(), time.Now(), "127.0.0.1", 0.0, 0.0))
				m.ExpectExec(regexp.QuoteMeta(repository.UserUpdateLastIPQuery)).
					WithArgs(ip, int64(2)).
					WillReturnError(errors.New("no session"))
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  true,
		},
		{
			name: "success",
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(repository.UserGetByLoginQuery)).
					WithArgs(req.Login).
					WillReturnRows(sqlmock.NewRows([]string{"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn"}).
						AddRow(int64(3), req.Login, hash, time.Now(), time.Now(), "127.0.0.1", 0.0, 0.0))
				m.ExpectQuery(regexp.QuoteMeta(repository.UserGetByIDQuery)).
					WithArgs(int64(3)).
					WillReturnRows(sqlmock.NewRows([]string{"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn"}).
						AddRow(int64(3), req.Login, hash, time.Now(), time.Now(), "127.0.0.1", 0.0, 0.0))
				m.ExpectExec(regexp.QuoteMeta(repository.UserUpdateLastIPQuery)).
					WithArgs(ip, int64(3)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec(regexp.QuoteMeta(repository.UserUpsertSessionQuery)).
					WithArgs(sqlmock.AnyArg(), int64(3), sqlmock.AnyArg(), ip).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantCode:  http.StatusOK,
			wantToken: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db, m, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock.New: %v", err)
			}
			defer db.Close()
			D := &conn.DB{DB: db}
			in := LoginInput{Req: req, IP: ip}
			if tt.name == "wrong password" {
				in.Req.Pass = "other"
			}
			tt.setup(m)
			s := NewService(D, repository.Repositories{User: repository.NewUserRepository(D), Order: repository.NewOrderRepository(D), Withdrawal: repository.NewWithdrawalRepository(D)})
			out := s.Login(ctx, in)
			if out.Code != tt.wantCode {
				t.Errorf("Code = %d, want %d", out.Code, tt.wantCode)
			}
			if tt.wantErr && out.Err == nil {
				t.Error("expected error")
			}
			if tt.wantToken && out.Token == "" {
				t.Error("expected non-empty token")
			}
			if err := m.ExpectationsWereMet(); err != nil {
				t.Fatalf("sqlmock expectations not met: %v", err)
			}
		})
	}
}
