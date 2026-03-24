package service

import (
	"errors"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/internal/gopherman/repository"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestService_GetBalance(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	uid := int64(9)

	tests := []struct {
		name     string
		setup    func(m sqlmock.Sqlmock)
		wantCode int
		want     model.BalanceResponse
		wantErr  bool
	}{
		{
			name: "GetByID fails",
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(repository.UserGetByIDQuery).
					WithArgs(uid).
					WillReturnError(errors.New("not found"))
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  true,
		},
		{
			name: "success",
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(repository.UserGetByIDQuery).
					WithArgs(uid).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn"}).
							AddRow(uid, "u", "p", time.Now(), time.Now(), "127.0.0.1", 500.5, 100.0),
					)
			},
			wantCode: http.StatusOK,
			want: model.BalanceResponse{
				Current:   500.5,
				Withdrawn: 100,
			},
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
			D := &conn.DB{DB: db}
			u := repository.NewUserRepository(D)
			o := repository.NewOrderRepository(D)
			w := repository.NewWithdrawalRepository(D)
			tt.setup(m)
			s := NewService(D, repository.Repositories{User: u, Order: o, Withdrawal: w})
			out := s.GetBalance(ctx, BalanceInput{UserID: uid})
			if out.Code != tt.wantCode {
				t.Errorf("Code = %d", out.Code)
			}
			if tt.wantErr && out.Err == nil {
				t.Error("want err")
			}
			if !tt.wantErr {
				if out.Balance.Current != tt.want.Current || out.Balance.Withdrawn != tt.want.Withdrawn {
					t.Errorf("Balance = %+v, want %+v", out.Balance, tt.want)
				}
			}
			if err := m.ExpectationsWereMet(); err != nil {
				t.Fatalf("sqlmock expectations not met: %v", err)
			}
		})
	}
}
