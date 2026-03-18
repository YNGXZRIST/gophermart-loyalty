package middleware

import (
	"database/sql"
	"errors"
	"gophermart-loyalty/internal/gopherman/contextkey"
	"gophermart-loyalty/internal/gopherman/handler/api"
	"gophermart-loyalty/internal/gopherman/repository"
	"gophermart-loyalty/internal/gopherman/repository/mock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
)

func TestAuthenticate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		authHeader     string
		setupMock      func(u *mock.MockUserRepository)
		wantCode       int
		wantUserID     int64
		wantNextCalled bool
	}{
		{
			name:           "no Authorization header",
			authHeader:     "",
			setupMock:      func(u *mock.MockUserRepository) {},
			wantCode:       http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "Bearer without token (empty after trim)",
			authHeader:     "Bearer ",
			setupMock:      func(u *mock.MockUserRepository) {},
			wantCode:       http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:       "valid Bearer token",
			authHeader: "Bearer session-abc",
			setupMock: func(u *mock.MockUserRepository) {
				u.EXPECT().UserIDFromSession(gomock.Any(), "session-abc").Return(int64(42), nil)
			},
			wantCode:       http.StatusOK,
			wantUserID:     42,
			wantNextCalled: true,
		},
		{
			name:       "token without Bearer prefix",
			authHeader: "raw-token",
			setupMock: func(u *mock.MockUserRepository) {
				u.EXPECT().UserIDFromSession(gomock.Any(), "raw-token").Return(int64(7), nil)
			},
			wantCode:       http.StatusOK,
			wantUserID:     7,
			wantNextCalled: true,
		},
		{
			name:       "session not found (sql.ErrNoRows)",
			authHeader: "Bearer unknown",
			setupMock: func(u *mock.MockUserRepository) {
				u.EXPECT().UserIDFromSession(gomock.Any(), "unknown").Return(int64(0), sql.ErrNoRows)
			},
			wantCode:       http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:       "repository error",
			authHeader: "Bearer x",
			setupMock: func(u *mock.MockUserRepository) {
				u.EXPECT().UserIDFromSession(gomock.Any(), "x").Return(int64(0), errors.New("db error"))
			},
			wantCode:       http.StatusInternalServerError,
			wantNextCalled: false,
		},
		{
			name:       "user id 0 with nil error",
			authHeader: "Bearer expired",
			setupMock: func(u *mock.MockUserRepository) {
				u.EXPECT().UserIDFromSession(gomock.Any(), "expired").Return(int64(0), nil)
			},
			wantCode:       http.StatusUnauthorized,
			wantNextCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRep := mock.NewMockUserRepository(ctrl)
			tt.setupMock(userRep)

			handler := api.NewHandler(nil, repository.Repositories{User: userRep}, zap.NewNop())

			var nextCalled bool
			var gotUserID int64
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				id, ok := contextkey.UserIDFromContext(r.Context())
				if ok {
					gotUserID = id
				}
				w.WriteHeader(http.StatusOK)
			})

			mw := Authenticate(handler)(next)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rr := httptest.NewRecorder()
			mw.ServeHTTP(rr, req)

			if rr.Code != tt.wantCode {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantCode)
			}
			if nextCalled != tt.wantNextCalled {
				t.Errorf("next called = %v, want %v", nextCalled, tt.wantNextCalled)
			}
			if tt.wantNextCalled && gotUserID != tt.wantUserID {
				t.Errorf("context userID = %d, want %d", gotUserID, tt.wantUserID)
			}
		})
	}
}
