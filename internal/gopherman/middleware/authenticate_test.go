package middleware

import (
	"database/sql"
	"errors"
	"gophermart-loyalty/internal/gopherman/contextkey"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/handler/api"
	"gophermart-loyalty/internal/gopherman/repository"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

func TestAuthenticate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		authHeader     string
		setupMock      func(m sqlmock.Sqlmock, rawToken string)
		wantCode       int
		wantUserID     int64
		wantNextCalled bool
	}{
		{
			name:           "no Authorization header",
			authHeader:     "",
			setupMock:      func(m sqlmock.Sqlmock, rawToken string) {},
			wantCode:       http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "Bearer without token (empty after trim)",
			authHeader:     "Bearer ",
			setupMock:      func(m sqlmock.Sqlmock, rawToken string) {},
			wantCode:       http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:       "valid Bearer token",
			authHeader: "Bearer session-abc",
			setupMock: func(m sqlmock.Sqlmock, rawToken string) {
				m.ExpectQuery(repository.UserUserIDFromSessionQuery).
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(
						sqlmock.NewRows([]string{"user_id", "expires_at", "ip", "created_at"}).
							AddRow(int64(42), time.Now().Add(time.Hour), "1.2.3.4", time.Now()),
					)
			},
			wantCode:       http.StatusOK,
			wantUserID:     42,
			wantNextCalled: true,
		},
		{
			name:       "token without Bearer prefix",
			authHeader: "raw-token",
			setupMock: func(m sqlmock.Sqlmock, rawToken string) {
				m.ExpectQuery(repository.UserUserIDFromSessionQuery).
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(
						sqlmock.NewRows([]string{"user_id", "expires_at", "ip", "created_at"}).
							AddRow(int64(7), time.Now().Add(time.Hour), "1.2.3.4", time.Now()),
					)
			},
			wantCode:       http.StatusOK,
			wantUserID:     7,
			wantNextCalled: true,
		},
		{
			name:       "session not found (sql.ErrNoRows)",
			authHeader: "Bearer unknown",
			setupMock: func(m sqlmock.Sqlmock, rawToken string) {
				m.ExpectQuery(repository.UserUserIDFromSessionQuery).
					WithArgs(sqlmock.AnyArg()).
					WillReturnError(sql.ErrNoRows)
			},
			wantCode:       http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:       "repository error",
			authHeader: "Bearer x",
			setupMock: func(m sqlmock.Sqlmock, rawToken string) {
				m.ExpectQuery(repository.UserUserIDFromSessionQuery).
					WithArgs(sqlmock.AnyArg()).
					WillReturnError(errors.New("db error"))
			},
			wantCode:       http.StatusInternalServerError,
			wantNextCalled: false,
		},
		{
			name:       "user id 0 with nil error",
			authHeader: "Bearer expired",
			setupMock: func(m sqlmock.Sqlmock, rawToken string) {
				m.ExpectQuery(repository.UserUserIDFromSessionQuery).
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(
						sqlmock.NewRows([]string{"user_id", "expires_at", "ip", "created_at"}).
							AddRow(int64(0), time.Now().Add(time.Hour), "1.2.3.4", time.Now()),
					)
			},
			wantCode:       http.StatusUnauthorized,
			wantNextCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db, m, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("sqlmock.New: %v", err)
			}
			defer db.Close()

			userRep := repository.NewUserRepository(&conn.DB{DB: db})
			tt.setupMock(m, tt.authHeader)
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
			if err := m.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock expectations not met: %v", err)
			}
		})
	}
}
