package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"gophermart-loyalty/internal/gopherman/auth/password"
	"gophermart-loyalty/internal/gopherman/auth/session"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/pkg/storage"
	"time"
)

type UserRepository interface {
	GetByLogin(ctx context.Context, login string) (*model.User, error)
	Register(ctx context.Context, login, pass, ip string) (*model.User, error)
	CreateSession(ctx context.Context, uid int64, ip string) (string, error)
	IsValidSession(ctx context.Context, token string) (bool, error)
}

type userRepo struct {
	db       *conn.DB
	users    *storage.MemStorage[string, *model.User]
	sessions *storage.MemStorage[string, *model.Sessions]
}

func NewUserRepository(db *conn.DB) UserRepository {
	users := storage.NewMemStorage[string, *model.User]()
	sessStorage := storage.NewMemStorage[string, *model.Sessions]()
	return &userRepo{db: db, users: users, sessions: sessStorage}
}

func (r *userRepo) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	u, err := r.users.Get(ctx, login)
	if err == nil {
		return u, nil
	}
	var dbUser model.User
	var lastIP sql.NullString
	err = r.db.DB.QueryRowContext(ctx,
		"SELECT id, login, pass, created_at, updated_at, last_login_ip FROM users WHERE login=$1",
		login,
	).Scan(&dbUser.ID, &dbUser.Login, &dbUser.Pass, &dbUser.CreatedAt, &dbUser.UpdatedAt, &lastIP)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if lastIP.Valid {
		dbUser.LastIp = lastIP.String
	}
	_ = r.users.Set(ctx, login, &dbUser)
	return &dbUser, nil
}

func (r *userRepo) Register(ctx context.Context, login, pass, ip string) (*model.User, error) {
	hash, err := password.Hash(pass)
	if err != nil {
		return nil, fmt.Errorf("could not hash password: %w", err)
	}
	var u model.User
	var lastIP sql.NullString
	err = r.db.QueryRowContext(ctx,
		"INSERT INTO users(login, pass, last_login_ip) VALUES ($1, $2, $3) RETURNING id, login, pass, created_at, updated_at, last_login_ip",
		login, hash, ip,
	).Scan(&u.ID, &u.Login, &u.Pass, &u.CreatedAt, &u.UpdatedAt, &lastIP)
	if err != nil {
		return nil, fmt.Errorf("could not insert user: %w", err)
	}
	if lastIP.Valid {
		u.LastIp = lastIP.String
	}
	_ = r.users.Set(ctx, login, &u)

	return &u, nil
}
func (r *userRepo) UpdateLastIp(ctx context.Context, uid int64, ip string) error {
	_, err := r.db.QueryContext(ctx, "UPDATE users SET last_login_ip=$1 where id=$2", ip, uid)
	return err
}
func (r *userRepo) CreateSession(ctx context.Context, uid int64, ip string) (string, error) {
	token, err := session.GenerateToken()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])
	expiresAt := time.Now().Add(24 * time.Hour)

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO sessions (token_hash, user_id, expires_at, ip) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (user_id, ip) DO UPDATE SET
		   token_hash = EXCLUDED.token_hash,
		   expires_at = EXCLUDED.expires_at`,
		tokenHash, uid, expiresAt, ip)
	if err != nil {
		return "", fmt.Errorf("upsert session: %w", err)
	}
	_ = r.sessions.Set(ctx, tokenHash, &model.Sessions{UserID: uid, TokenHash: tokenHash, ExpiresAt: expiresAt, IP: ip})
	return token, nil
}

func (r *userRepo) IsValidSession(ctx context.Context, token string) (bool, error) {
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])
	cache, err := r.sessions.Get(ctx, tokenHash)
	if err == nil {
		if !cache.ExpiresAt.After(time.Now()) {
			return false, nil
		}
		return true, nil
	}
	var dbSession model.Sessions
	var ipNull sql.NullString
	err = r.db.DB.QueryRowContext(ctx,
		"SELECT user_id, expires_at, ip, created_at FROM sessions WHERE token_hash = $1 AND expires_at > CURRENT_TIMESTAMP",
		tokenHash).Scan(&dbSession.UserID, &dbSession.ExpiresAt, &ipNull, &dbSession.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	if ipNull.Valid {
		dbSession.IP = ipNull.String
	}
	dbSession.TokenHash = tokenHash
	_ = r.sessions.Set(ctx, tokenHash, &dbSession)
	return true, nil
}
