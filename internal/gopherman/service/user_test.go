package service

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"

	"gophermart-loyalty/internal/gopherman/auth/password"
	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/internal/gopherman/repository"
	"gophermart-loyalty/internal/gopherman/repository/mock"

	"github.com/golang/mock/gomock"
)

func TestService_Register(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	req := model.RegisterRequest{Login: "user1", Pass: "secret12"}
	ip := "127.0.0.1"

	tests := []struct {
		name       string
		setup      func(u *mock.MockUserRepository)
		wantCode   int
		wantErr    bool
		wantToken  bool
		tokenValue string
	}{
		{
			name: "GetByLogin database error",
			setup: func(u *mock.MockUserRepository) {
				u.EXPECT().GetByLogin(ctx, req.Login).Return(nil, errors.New("db down"))
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  true,
		},
		{
			name: "login already taken",
			setup: func(u *mock.MockUserRepository) {
				u.EXPECT().GetByLogin(ctx, req.Login).Return(&model.User{ID: 1, Login: req.Login}, nil)
			},
			wantCode: http.StatusConflict,
			wantErr:  true,
		},
		{
			name: "new user ErrNoRows then register fails",
			setup: func(u *mock.MockUserRepository) {
				u.EXPECT().GetByLogin(ctx, req.Login).Return(nil, sql.ErrNoRows)
				u.EXPECT().Register(ctx, req.Login, req.Pass, ip).Return(nil, errors.New("insert failed"))
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  true,
		},
		{
			name: "session creation fails",
			setup: func(u *mock.MockUserRepository) {
				u.EXPECT().GetByLogin(ctx, req.Login).Return(nil, sql.ErrNoRows)
				u.EXPECT().Register(ctx, req.Login, req.Pass, ip).Return(&model.User{ID: 42, Login: req.Login}, nil)
				u.EXPECT().CreateSession(ctx, int64(42), ip).Return("", errors.New("session error"))
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  true,
		},
		{
			name: "success",
			setup: func(u *mock.MockUserRepository) {
				u.EXPECT().GetByLogin(ctx, req.Login).Return(nil, sql.ErrNoRows)
				u.EXPECT().Register(ctx, req.Login, req.Pass, ip).Return(&model.User{ID: 7, Login: req.Login}, nil)
				u.EXPECT().CreateSession(ctx, int64(7), ip).Return("jwt-token-xyz", nil)
			},
			wantCode:   http.StatusOK,
			wantToken:  true,
			tokenValue: "jwt-token-xyz",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			u := mock.NewMockUserRepository(ctrl)
			o := mock.NewMockOrderRepository(ctrl)
			w := mock.NewMockWithdrawalRepository(ctrl)
			tt.setup(u)
			s := NewService(nil, repository.Repositories{User: u, Order: o, Withdrawal: w})
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
			if tt.wantToken && out.Token != tt.tokenValue {
				t.Errorf("Token = %q, want %q", out.Token, tt.tokenValue)
			}
		})
	}
}

func TestService_Login(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	req := model.RegisterRequest{Login: "alice", Pass: "correcthorse"}
	ip := "10.0.0.1"
	hash, err := password.Hash(req.Pass)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		setup      func(u *mock.MockUserRepository)
		wantCode   int
		wantErr    bool
		wantToken  bool
		tokenValue string
	}{
		{
			name: "GetByLogin error",
			setup: func(u *mock.MockUserRepository) {
				u.EXPECT().GetByLogin(ctx, req.Login).Return(nil, errors.New("timeout"))
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  true,
		},
		{
			name: "user not found",
			setup: func(u *mock.MockUserRepository) {
				u.EXPECT().GetByLogin(ctx, req.Login).Return(nil, nil)
			},
			wantCode: http.StatusUnauthorized,
			wantErr:  true,
		},
		{
			name: "wrong password",
			setup: func(u *mock.MockUserRepository) {
				u.EXPECT().GetByLogin(ctx, req.Login).Return(&model.User{ID: 1, Login: req.Login, Pass: hash}, nil)
			},
			wantCode: http.StatusUnauthorized,
			wantErr:  true,
		},
		{
			name: "CreateSession fails",
			setup: func(u *mock.MockUserRepository) {
				u.EXPECT().GetByLogin(ctx, req.Login).Return(&model.User{ID: 2, Login: req.Login, Pass: hash}, nil)
				u.EXPECT().CreateSession(ctx, int64(2), ip).Return("", errors.New("no session"))
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  true,
		},
		{
			name: "success",
			setup: func(u *mock.MockUserRepository) {
				u.EXPECT().GetByLogin(ctx, req.Login).Return(&model.User{ID: 3, Login: req.Login, Pass: hash}, nil)
				u.EXPECT().CreateSession(ctx, int64(3), ip).Return("sess-abc", nil)
			},
			wantCode:   http.StatusOK,
			wantToken:  true,
			tokenValue: "sess-abc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			u := mock.NewMockUserRepository(ctrl)
			o := mock.NewMockOrderRepository(ctrl)
			w := mock.NewMockWithdrawalRepository(ctrl)
			in := LoginInput{Req: req, IP: ip}
			if tt.name == "wrong password" {
				in.Req.Pass = "other"
			}
			tt.setup(u)
			s := NewService(nil, repository.Repositories{User: u, Order: o, Withdrawal: w})
			out := s.Login(ctx, in)
			if out.Code != tt.wantCode {
				t.Errorf("Code = %d, want %d", out.Code, tt.wantCode)
			}
			if tt.wantErr && out.Err == nil {
				t.Error("expected error")
			}
			if tt.wantToken && out.Token != tt.tokenValue {
				t.Errorf("Token = %q", out.Token)
			}
		})
	}
}
