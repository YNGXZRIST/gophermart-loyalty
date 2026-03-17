package api

import (
	"bytes"
	"context"
	"gophermart-loyalty/internal/gopherman/auth/password"
	"gophermart-loyalty/internal/gopherman/constant"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/internal/gopherman/repository"
	"gophermart-loyalty/internal/gopherman/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

var _ repository.UserRepository = (*mockUserRepo)(nil)

type mockUserRepo struct {
	getByLogin        func(ctx context.Context, login string) (*model.User, error)
	getByID           func(ctx context.Context, id int64) (*model.User, error)
	register          func(ctx context.Context, login, pass, ip string) (*model.User, error)
	createSession     func(ctx context.Context, uid int64, ip string) (string, error)
	userIDFromSession func(ctx context.Context, token string) (int64, error)
}

func (m *mockUserRepo) IncrementBalance(_ context.Context, _ *conn.Tx, _ int64, _ float64) error {
	return nil
}

func (m *mockUserRepo) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	if m.getByLogin != nil {
		return m.getByLogin(ctx, login)
	}
	return nil, nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	return nil, nil
}

func (m *mockUserRepo) Register(ctx context.Context, login, pass, ip string) (*model.User, error) {
	if m.register != nil {
		return m.register(ctx, login, pass, ip)
	}
	return nil, nil
}

func (m *mockUserRepo) CreateSession(ctx context.Context, uid int64, ip string) (string, error) {
	if m.createSession != nil {
		return m.createSession(ctx, uid, ip)
	}
	return "", nil
}

func (m *mockUserRepo) UserIDFromSession(ctx context.Context, token string) (int64, error) {
	if m.userIDFromSession != nil {
		return m.userIDFromSession(ctx, token)
	}
	return 0, nil
}

func (m *mockUserRepo) IncrementWithdrawn(ctx context.Context, tx *conn.Tx, w *model.Withdrawal) error {
	return nil
}

func TestHandler_Register(t *testing.T) {
	lgr := zap.NewNop()

	tests := []struct {
		name           string
		body           string
		contentType    string
		mock           *mockUserRepo
		wantStatus     int
		wantAuthHeader bool
	}{
		{
			name:        "invalid JSON returns 400",
			body:        `{`,
			contentType: constant.ApplicationJSON,
			mock:        &mockUserRepo{},
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "invalid body validation - short login",
			body:        `{"login":"ab","password":"password123"}`,
			contentType: constant.ApplicationJSON,
			mock:        &mockUserRepo{},
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "invalid body validation - short password",
			body:        `{"login":"user123","password":"1234"}`,
			contentType: constant.ApplicationJSON,
			mock:        &mockUserRepo{},
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "user already exists returns 409",
			body:        `{"login":"existing","password":"password123"}`,
			contentType: constant.ApplicationJSON,
			mock: &mockUserRepo{
				getByLogin: func(ctx context.Context, login string) (*model.User, error) {
					return &model.User{ID: 1, Login: login}, nil
				},
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:        "GetByLogin error returns 500",
			body:        `{"login":"user123","password":"password123"}`,
			contentType: constant.ApplicationJSON,
			mock: &mockUserRepo{
				getByLogin: func(ctx context.Context, login string) (*model.User, error) {
					return nil, context.DeadlineExceeded
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:        "Register error returns 500",
			body:        `{"login":"user123","password":"password123"}`,
			contentType: "application/json",
			mock: &mockUserRepo{
				getByLogin: func(ctx context.Context, login string) (*model.User, error) {
					return nil, nil
				},
				register: func(ctx context.Context, login, pass, ip string) (*model.User, error) {
					return nil, context.DeadlineExceeded
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:        "CreateSession error returns 500",
			body:        `{"login":"user123","password":"password123"}`,
			contentType: constant.ApplicationJSON,
			mock: &mockUserRepo{
				getByLogin: func(ctx context.Context, login string) (*model.User, error) {
					return nil, nil
				},
				register: func(ctx context.Context, login, pass, ip string) (*model.User, error) {
					return &model.User{ID: 1, Login: login}, nil
				},
				createSession: func(ctx context.Context, uid int64, ip string) (string, error) {
					return "", context.DeadlineExceeded
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:        "success returns 200 and sets Authorization",
			body:        `{"login":"newuser","password":"password123"}`,
			contentType: "application/json",
			mock: &mockUserRepo{
				getByLogin: func(ctx context.Context, login string) (*model.User, error) {
					return nil, nil
				},
				register: func(ctx context.Context, login, pass, ip string) (*model.User, error) {
					return &model.User{ID: 42, Login: login}, nil
				},
				createSession: func(ctx context.Context, uid int64, ip string) (string, error) {
					return "token123", nil
				},
			},
			wantStatus:     http.StatusOK,
			wantAuthHeader: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repos := repository.Repositories{User: tt.mock, Order: nil, Withdrawal: nil}
			svc := service.NewService(nil, repos)
			h := &Handler{ser: svc, lgr: lgr}
			req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			req.RemoteAddr = "192.168.1.1:12345"
			rec := httptest.NewRecorder()

			h.Register(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantAuthHeader {
				auth := rec.Header().Get("Authorization")
				if auth != "Bearer token123" {
					t.Errorf("Authorization = %q, want %q", auth, "Bearer token123")
				}
			}
		})
	}
}

func TestHandler_Login(t *testing.T) {
	lgr := zap.NewNop()

	tests := []struct {
		name           string
		body           string
		contentType    string
		mock           *mockUserRepo
		wantStatus     int
		wantAuthHeader bool
	}{
		{
			name:        "invalid JSON returns 400",
			body:        `{`,
			contentType: "application/json",
			mock:        &mockUserRepo{},
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "invalid validation returns 400",
			body:        `{"login":"ab","password":"x"}`,
			contentType: "application/json",
			mock:        &mockUserRepo{},
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "user not found returns 401",
			body:        `{"login":"nobody","password":"password123"}`,
			contentType: "application/json",
			mock: &mockUserRepo{
				getByLogin: func(ctx context.Context, login string) (*model.User, error) {
					return nil, nil
				},
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:        "wrong password returns 401",
			body:        `{"login":"user","password":"wrongpass"}`,
			contentType: "application/json",
			mock: &mockUserRepo{
				getByLogin: func(ctx context.Context, login string) (*model.User, error) {
					hash, _ := password.Hash("rightpass")
					return &model.User{ID: 1, Login: "user", Pass: hash}, nil
				},
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:        "GetByLogin error returns 500",
			body:        `{"login":"user","password":"password123"}`,
			contentType: "application/json",
			mock: &mockUserRepo{
				getByLogin: func(ctx context.Context, login string) (*model.User, error) {
					return nil, context.DeadlineExceeded
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:        "CreateSession error returns 500",
			body:        `{"login":"user","password":"password123"}`,
			contentType: "application/json",
			mock: &mockUserRepo{
				getByLogin: func(ctx context.Context, login string) (*model.User, error) {
					hash, _ := password.Hash("password123")
					return &model.User{ID: 1, Login: "user", Pass: hash}, nil
				},
				createSession: func(ctx context.Context, uid int64, ip string) (string, error) {
					return "", context.DeadlineExceeded
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:        "success returns 200 and sets Authorization",
			body:        `{"login":"user","password":"password123"}`,
			contentType: "application/json",
			mock: &mockUserRepo{
				getByLogin: func(ctx context.Context, login string) (*model.User, error) {
					hash, _ := password.Hash("password123")
					return &model.User{ID: 1, Login: "user", Pass: hash}, nil
				},
				createSession: func(ctx context.Context, uid int64, ip string) (string, error) {
					return "session-token-xyz", nil
				},
			},
			wantStatus:     http.StatusOK,
			wantAuthHeader: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repos := repository.Repositories{User: tt.mock, Order: nil, Withdrawal: nil}
			svc := service.NewService(nil, repos)
			h := &Handler{ser: svc, lgr: lgr}
			req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			req.RemoteAddr = "192.168.1.1:12345"
			rec := httptest.NewRecorder()

			h.Login(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantAuthHeader {
				auth := rec.Header().Get("Authorization")
				if auth == "" || len(auth) < 8 || auth[:7] != "Bearer " {
					t.Errorf("Authorization should start with Bearer , got %q", auth)
				}
			}
		})
	}
}
