package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/db/trmanager"
	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/internal/gopherman/repository/mock"
	"gophermart-loyalty/pkg/storage"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
)

func TestUserRepository_WithMocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUser := mock.NewMockUserRepository(ctrl)

	repos := Repositories{
		User: mockUser,
	}
	_ = repos
}

func Test_userRepo_GetByLogin(t *testing.T) {
	ctx := context.Background()
	login := "test"
	userID := int64(1)
	pass := "pass"
	lastIP := "1.2.3.4"
	balance := 123.45
	withdrawn := 10.5

	db, mockSQL := newMockConnDB(t)
	r := NewUserRepository(db).(*userRepo)
	mockSQL.MatchExpectationsInOrder(false)

	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now().Add(-time.Minute)

	mockSQL.ExpectQuery(userGetByLoginQuery).
		WithArgs(login).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn"}).
				AddRow(userID, login, pass, createdAt, updatedAt, lastIP, balance, withdrawn),
		)

	got, err := r.GetByLogin(ctx, login)
	if err != nil {
		t.Fatalf("GetByLogin() error = %v", err)
	}
	if got.ID != userID || got.Login != login || got.LastIP != lastIP || got.Balance != balance || got.Withdrawn != withdrawn {
		t.Fatalf("GetByLogin() got %+v, want id=%d login=%s lastIP=%s", got, userID, login, lastIP)
	}

	cachedID, err := r.loginToID.Get(ctx, login)
	if err != nil {
		t.Fatalf("loginToID.Get() error = %v", err)
	}
	if cachedID != userID {
		t.Fatalf("loginToID.Get() = %d, want %d", cachedID, userID)
	}
	cachedUser, err := r.usersByID.Get(ctx, userID)
	if err != nil {
		t.Fatalf("usersByID.Get() error = %v", err)
	}
	if cachedUser.ID != userID || cachedUser.Login != login {
		t.Fatalf("usersByID.Get() got %+v, want id=%d login=%s", cachedUser, userID, login)
	}

	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}

func Test_userRepo_GetByID(t *testing.T) {
	ctx := context.Background()
	userID := int64(42)
	login := "bob"
	pass := "pass"
	lastIP := "5.6.7.8"
	balance := 250.0
	withdrawn := 35.5

	db, mockSQL := newMockConnDB(t)
	r := NewUserRepository(db).(*userRepo)
	mockSQL.MatchExpectationsInOrder(false)

	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now().Add(-time.Minute)

	mockSQL.ExpectQuery(UserGetByIDQuery).
		WithArgs(userID).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn"}).
				AddRow(userID, login, pass, createdAt, updatedAt, lastIP, balance, withdrawn),
		)

	got, err := r.GetByID(ctx, userID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.ID != userID || got.Login != login || got.LastIP != lastIP || got.Balance != balance || got.Withdrawn != withdrawn {
		t.Fatalf("GetByID() got %+v, want id=%d login=%s lastIP=%s", got, userID, login, lastIP)
	}

	cachedID, err := r.loginToID.Get(ctx, login)
	if err != nil {
		t.Fatalf("loginToID.Get() error = %v", err)
	}
	if cachedID != userID {
		t.Fatalf("loginToID.Get() = %d, want %d", cachedID, userID)
	}
	cachedUser, err := r.usersByID.Get(ctx, userID)
	if err != nil {
		t.Fatalf("usersByID.Get() error = %v", err)
	}
	if cachedUser.ID != userID || cachedUser.Login != login {
		t.Fatalf("usersByID.Get() got %+v, want id=%d login=%s", cachedUser, userID, login)
	}

	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}

func Test_userRepo_Register(t *testing.T) {
	ctx := context.Background()
	login := "newuser"
	pass := "secret"
	ip := "1.2.3.4"
	userID := int64(3)

	db, mockSQL := newMockConnDB(t)
	r := NewUserRepository(db).(*userRepo)
	mockSQL.MatchExpectationsInOrder(false)

	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now().Add(-time.Minute)

	mockSQL.ExpectQuery(userRegisterQuery).
		WithArgs(login, sqlmock.AnyArg(), ip).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "login", "pass", "created_at", "updated_at", "last_login_ip"}).
				AddRow(userID, login, "returned-hash", createdAt, updatedAt, ip),
		)

	got, err := r.Register(ctx, login, pass, ip)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if got.ID != userID || got.Login != login || got.Pass != "returned-hash" || got.LastIP != ip {
		t.Fatalf("Register() got %+v, want id=%d login=%s passHash=%s lastIP=%s", got, userID, login, "returned-hash", ip)
	}

	cachedID, err := r.loginToID.Get(ctx, login)
	if err != nil {
		t.Fatalf("loginToID.Get() error = %v", err)
	}
	if cachedID != userID {
		t.Fatalf("loginToID.Get() = %d, want %d", cachedID, userID)
	}
	cachedUser, err := r.usersByID.Get(ctx, userID)
	if err != nil {
		t.Fatalf("usersByID.Get() error = %v", err)
	}
	if cachedUser.ID != userID || cachedUser.Login != login {
		t.Fatalf("usersByID.Get() got %+v, want id=%d login=%s", cachedUser, userID, login)
	}

	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}

func TestNewUserRepository(t *testing.T) {
	db := &conn.DB{}
	got := NewUserRepository(db)
	if got == nil {
		t.Fatalf("NewUserRepository returned nil")
	}
	if _, ok := got.(*userRepo); !ok {
		t.Fatalf("NewUserRepository() type = %T, want *userRepo", got)
	}
}

func Test_userRepo_CreateSession(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)
	ip := "1.2.3.4"

	db, mockSQL := newMockConnDB(t)
	r := NewUserRepository(db).(*userRepo)
	mockSQL.MatchExpectationsInOrder(false)

	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now().Add(-time.Minute)

	mockSQL.ExpectQuery(UserGetByIDQuery).
		WithArgs(userID).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn"}).
				AddRow(userID, "test", "pass", createdAt, updatedAt, "old-ip", 0.0, 0.0),
		)

	mockSQL.ExpectExec(userUpdateLastIPQuery).
		WithArgs(ip, userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mockSQL.ExpectExec(userUpsertSessionQuery).
		WithArgs(sqlmock.AnyArg(), userID, sqlmock.AnyArg(), ip).
		WillReturnResult(sqlmock.NewResult(1, 1))

	token, err := r.CreateSession(ctx, userID, ip)
	if err != nil {
		t.Fatalf("CreateSession() error = %v, want nil", err)
	}
	if token == "" {
		t.Fatalf("CreateSession() returned empty token")
	}

	sum := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(sum[:])

	cached, err := r.sessions.Get(ctx, tokenHash)
	if err != nil {
		t.Fatalf("sessions.Get() error = %v", err)
	}
	if cached.UserID != userID || cached.TokenHash != tokenHash || cached.IP != ip {
		t.Fatalf("cached session got %+v, want userID=%d tokenHash=%s ip=%s", cached, userID, tokenHash, ip)
	}
	if !cached.ExpiresAt.After(time.Now()) {
		t.Fatalf("cached.ExpiresAt=%v must be in the future", cached.ExpiresAt)
	}

	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}

func Test_userRepo_IncrementBalance(t *testing.T) {
	ctx := context.Background()

	t.Run("zero_increment_no_db_calls", func(t *testing.T) {
		r := &userRepo{}
		if err := r.IncrementBalance(ctx, 123, 0); err != nil {
			t.Fatalf("IncrementBalance() error = %v, want nil", err)
		}
	})

	t.Run("increment_updates_balance_and_cache", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			t.Fatalf("sqlmock.New: %v", err)
		}
		mock.MatchExpectationsInOrder(false)
		defer db.Close()

		sqlDB := &conn.DB{DB: db}
		r := &userRepo{
			repoBase:  repoBase{db: sqlDB},
			loginToID: storage.NewMemStorage[string, int64](),
			usersByID: storage.NewMemStorage[int64, *model.User](),
			sessions:  storage.NewMemStorage[string, *model.Sessions](),
		}

		userID := int64(1)
		increment := 10.0
		initialBalance := 100.5
		expectedBalance := initialBalance + increment
		createdAt := time.Now().Add(-time.Hour)
		updatedAt := time.Now().Add(-time.Minute)

		mock.ExpectQuery(UserGetByIDQuery).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn",
			}).AddRow(userID, "test", "pass", createdAt, updatedAt, "old-ip", initialBalance, 20.0))

		mock.ExpectBegin()
		txSQL, err := db.Begin()
		if err != nil {
			t.Fatalf("db.Begin: %v", err)
		}
		defer func() { _ = txSQL.Rollback() }()

		mock.ExpectExec(userIncrementBalanceQuery).
			WithArgs(expectedBalance, userID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		txCtx := trmanager.WithTx(ctx, &trmanager.Tx{Tx: txSQL})
		if err := r.IncrementBalance(txCtx, userID, increment); err != nil {
			t.Fatalf("IncrementBalance() error = %v, want nil", err)
		}

		gotUser, err := r.usersByID.Get(ctx, userID)
		if err != nil {
			t.Fatalf("cache Get error = %v", err)
		}
		if gotUser.Balance != expectedBalance {
			t.Fatalf("cache balance = %v, want %v", gotUser.Balance, expectedBalance)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("sqlmock expectations not met: %v", err)
		}
	})
}

func Test_userRepo_IncrementWithdrawn(t *testing.T) {
	ctx := context.Background()

	t.Run("decreases_balance_and_increases_withdrawn_and_cache", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			t.Fatalf("sqlmock.New: %v", err)
		}
		mock.MatchExpectationsInOrder(false)
		defer db.Close()

		sqlDB := &conn.DB{DB: db}
		r := &userRepo{
			repoBase:  repoBase{db: sqlDB},
			loginToID: storage.NewMemStorage[string, int64](),
			usersByID: storage.NewMemStorage[int64, *model.User](),
			sessions:  storage.NewMemStorage[string, *model.Sessions](),
		}

		userID := int64(1)
		initialBalance := 100.0
		initialWithdrawn := 20.0
		sum := 25.0
		expectedBalance := initialBalance - sum
		expectedWithdrawn := initialWithdrawn + sum
		createdAt := time.Now().Add(-time.Hour)
		updatedAt := time.Now().Add(-time.Minute)

		mock.ExpectQuery(UserGetByIDQuery).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn",
			}).AddRow(userID, "test", "pass", createdAt, updatedAt, "old-ip", initialBalance, initialWithdrawn))

		mock.ExpectBegin()
		txSQL, err := db.Begin()
		if err != nil {
			t.Fatalf("db.Begin: %v", err)
		}
		defer func() { _ = txSQL.Rollback() }()

		mock.ExpectExec(userIncrementWithdrawnQuery).
			WithArgs(expectedBalance, expectedWithdrawn, userID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		txCtx := trmanager.WithTx(ctx, &trmanager.Tx{Tx: txSQL})
		if err := r.IncrementWithdrawn(txCtx, &model.Withdrawal{UserID: userID, OrderID: "w1", Sum: sum}); err != nil {
			t.Fatalf("IncrementWithdrawn() error = %v, want nil", err)
		}

		gotUser, err := r.usersByID.Get(ctx, userID)
		if err != nil {
			t.Fatalf("cache Get error = %v", err)
		}
		if gotUser.Balance != expectedBalance {
			t.Fatalf("cache balance = %v, want %v", gotUser.Balance, expectedBalance)
		}
		if gotUser.Withdrawn != expectedWithdrawn {
			t.Fatalf("cache withdrawn = %v, want %v", gotUser.Withdrawn, expectedWithdrawn)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("sqlmock expectations not met: %v", err)
		}
	})
}

func Test_userRepo_UpdateLastIP(t *testing.T) {
	ctx := context.Background()

	db, m, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	sqlDB := &conn.DB{DB: db}
	r := &userRepo{
		repoBase:  repoBase{db: sqlDB},
		loginToID: storage.NewMemStorage[string, int64](),
		usersByID: storage.NewMemStorage[int64, *model.User](),
		sessions:  storage.NewMemStorage[string, *model.Sessions](),
	}

	userID := int64(1)
	newIP := "1.2.3.4"
	initialBalance := 100.0
	initialWithdrawn := 20.0
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now().Add(-time.Minute)

	m.ExpectQuery(UserGetByIDQuery).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn",
		}).AddRow(userID, "test", "pass", createdAt, updatedAt, "old-ip", initialBalance, initialWithdrawn))

	m.ExpectExec(userUpdateLastIPQuery).
		WithArgs(newIP, userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := r.UpdateLastIP(ctx, userID, newIP); err != nil {
		t.Fatalf("UpdateLastIP() error = %v, want nil", err)
	}

	gotUser, err := r.usersByID.Get(ctx, userID)
	if err != nil {
		t.Fatalf("cache Get error = %v", err)
	}
	if gotUser.LastIP != newIP {
		t.Fatalf("cache LastIP = %q, want %q", gotUser.LastIP, newIP)
	}

	if err := m.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}

func Test_userRepo_UserIDFromSession(t *testing.T) {
	ctx := context.Background()
	token := "token-123"
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	t.Run("cache_hit_valid_and_not_expired", func(t *testing.T) {
		sessions := storage.NewMemStorage[string, *model.Sessions]()
		_ = sessions.Set(ctx, tokenHash, &model.Sessions{
			UserID:    42,
			TokenHash: tokenHash,
			IP:        "1.2.3.4",
			ExpiresAt: time.Now().Add(time.Hour),
		})

		r := &userRepo{
			sessions: sessions,
		}

		got, err := r.UserIDFromSession(ctx, token)
		if err != nil {
			t.Fatalf("UserIDFromSession() error = %v, want nil", err)
		}
		if got != 42 {
			t.Fatalf("UserIDFromSession() = %v, want 42", got)
		}
	})

	t.Run("cache_hit_expired_returns_zero_nil", func(t *testing.T) {
		sessions := storage.NewMemStorage[string, *model.Sessions]()
		_ = sessions.Set(ctx, tokenHash, &model.Sessions{
			UserID:    42,
			TokenHash: tokenHash,
			IP:        "1.2.3.4",
			ExpiresAt: time.Now().Add(-time.Hour),
		})

		r := &userRepo{
			sessions: sessions,
		}

		got, err := r.UserIDFromSession(ctx, token)
		if err != nil {
			t.Fatalf("UserIDFromSession() error = %v, want nil", err)
		}
		if got != 0 {
			t.Fatalf("UserIDFromSession() = %v, want 0", got)
		}
	})
}
