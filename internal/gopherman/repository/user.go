package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"gophermart-loyalty/internal/gopherman/auth/password"
	"gophermart-loyalty/internal/gopherman/auth/session"
	"gophermart-loyalty/internal/gopherman/constant"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/errors/labelerrors"
	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/pkg/storage"
	"time"
)

type UserRepository interface {
	GetByLogin(ctx context.Context, login string) (*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
	Register(ctx context.Context, login, pass, ip string) (*model.User, error)
	CreateSession(ctx context.Context, uid int64, ip string) (string, error)
	UserIDFromSession(ctx context.Context, token string) (int64, error)
	IncrementWithdrawn(ctx context.Context, w *model.Withdrawal) error
	IncrementBalance(ctx context.Context, userID int64, increment float64) error
}

type userRepo struct {
	repoBase  repoBase
	loginToID *storage.MemStorage[string, int64]
	usersByID *storage.MemStorage[int64, *model.User]
	sessions  *storage.MemStorage[string, *model.Sessions]
}

const (
	userGetByLoginQuery    = "SELECT id, login, pass, created_at, updated_at, last_login_ip, balance, withdrawn FROM users WHERE login=$1"
	UserGetByIDQuery       = "SELECT id, login, pass, created_at, updated_at, last_login_ip, balance, withdrawn FROM users WHERE id=$1"
	userRegisterQuery      = "INSERT INTO users(login, pass, last_login_ip) VALUES ($1, $2, $3) RETURNING id, login, pass, created_at, updated_at, last_login_ip"
	userUpdateLastIPQuery  = "UPDATE users SET last_login_ip=$1 where id=$2"
	userUpsertSessionQuery = `INSERT INTO sessions (token_hash, user_id, expires_at, ip) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (user_id, ip) DO UPDATE SET
		   token_hash = EXCLUDED.token_hash,
		   expires_at = EXCLUDED.expires_at`
	userUserIDFromSessionQuery = "SELECT user_id, expires_at, ip, created_at FROM sessions WHERE token_hash = $1 AND expires_at > CURRENT_TIMESTAMP"

	userIncrementWithdrawnQuery = `UPDATE users SET balance = $1,withdrawn = $2  WHERE id = $3`
	userIncrementBalanceQuery   = "UPDATE users SET balance = $1 WHERE id = $2;"
)

func NewUserRepository(db *conn.DB) UserRepository {
	loginToID := storage.NewMemStorage[string, int64]()
	usersByID := storage.NewMemStorage[int64, *model.User]()
	sessStorage := storage.NewMemStorage[string, *model.Sessions]()

	return &userRepo{repoBase: repoBase{db: db}, loginToID: loginToID, usersByID: usersByID, sessions: sessStorage}
}

func (r *userRepo) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	if id, err := r.loginToID.Get(ctx, login); err == nil {
		return r.GetByID(ctx, id)
	}
	var dbUser model.User
	var lastIP sql.NullString
	err := r.repoBase.q(ctx).QueryRowContext(ctx,
		userGetByLoginQuery,
		login,
	).Scan(&dbUser.ID, &dbUser.Login, &dbUser.Pass, &dbUser.CreatedAt, &dbUser.UpdatedAt, &lastIP, &dbUser.Balance, &dbUser.Withdrawn)
	if err != nil {
		return nil, labelerrors.NewLabelError(constant.LabelRepository+".User.GetByLogin.Query", err)
	}
	if lastIP.Valid {
		dbUser.LastIP = lastIP.String
	}
	_ = r.loginToID.Set(ctx, login, dbUser.ID)
	_ = r.usersByID.Set(ctx, dbUser.ID, &dbUser)
	return &dbUser, nil
}

func (r *userRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	u, err := r.usersByID.Get(ctx, id)
	if err == nil {
		return u, nil
	}
	var dbUser model.User
	var lastIP sql.NullString
	err = r.repoBase.q(ctx).QueryRowContext(ctx,
		UserGetByIDQuery,
		id,
	).Scan(&dbUser.ID, &dbUser.Login, &dbUser.Pass, &dbUser.CreatedAt, &dbUser.UpdatedAt, &lastIP, &dbUser.Balance, &dbUser.Withdrawn)
	if err != nil {
		return nil, labelerrors.NewLabelError(constant.LabelRepository+".User.GetByID.Query", err)
	}
	if lastIP.Valid {
		dbUser.LastIP = lastIP.String
	}
	_ = r.loginToID.Set(ctx, dbUser.Login, dbUser.ID)
	_ = r.usersByID.Set(ctx, dbUser.ID, &dbUser)
	return &dbUser, nil
}

func (r *userRepo) Register(ctx context.Context, login, pass, ip string) (*model.User, error) {
	hash, err := password.Hash(pass)
	if err != nil {
		return nil, labelerrors.NewLabelError(constant.LabelRepository+".User.Register.HashPassword", err)
	}
	var u model.User
	var lastIP sql.NullString
	err = r.repoBase.q(ctx).QueryRowContext(ctx,
		userRegisterQuery,
		login, hash, ip,
	).Scan(&u.ID, &u.Login, &u.Pass, &u.CreatedAt, &u.UpdatedAt, &lastIP)
	if err != nil {
		return nil, labelerrors.NewLabelError(constant.LabelRepository+".User.Register.Insert", err)
	}
	if lastIP.Valid {
		u.LastIP = lastIP.String
	}
	_ = r.loginToID.Set(ctx, login, u.ID)
	_ = r.usersByID.Set(ctx, u.ID, &u)
	return &u, nil
}
func (r *userRepo) UpdateLastIP(ctx context.Context, userID int64, ip string) error {
	u, err := r.GetByID(ctx, userID)
	if err != nil {
		return labelerrors.NewLabelError(constant.LabelRepository+".User.UpdateLastIP.GetByID", err)
	}
	_, err = r.repoBase.q(ctx).ExecContext(ctx, userUpdateLastIPQuery, ip, userID)
	if err != nil {
		return labelerrors.NewLabelError(constant.LabelRepository+".User.UpdateLastIP.Exec", err)
	}
	u.LastIP = ip
	return nil
}
func (r *userRepo) CreateSession(ctx context.Context, userID int64, ip string) (string, error) {
	token, err := session.GenerateToken()
	if err != nil {
		return "", labelerrors.NewLabelError(constant.LabelRepository+".User.CreateSession.GenerateToken", err)
	}
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])
	expiresAt := time.Now().Add(24 * time.Hour)
	err = r.UpdateLastIP(ctx, userID, ip)
	if err != nil {
		return "", err
	}
	_, err = r.repoBase.q(ctx).ExecContext(ctx,
		userUpsertSessionQuery,
		tokenHash, userID, expiresAt, ip)
	if err != nil {
		return "", labelerrors.NewLabelError(constant.LabelRepository+".User.CreateSession.Upsert", err)
	}
	_ = r.sessions.Set(ctx, tokenHash, &model.Sessions{UserID: userID, TokenHash: tokenHash, ExpiresAt: expiresAt, IP: ip})
	return token, nil
}

func (r *userRepo) UserIDFromSession(ctx context.Context, token string) (int64, error) {
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])
	cache, err := r.sessions.Get(ctx, tokenHash)
	if err == nil {
		if !cache.ExpiresAt.After(time.Now()) {
			return 0, nil
		}
		return cache.UserID, nil
	}
	var dbSession model.Sessions
	var ipNull sql.NullString
	err = r.repoBase.q(ctx).QueryRowContext(ctx,
		userUserIDFromSessionQuery,
		tokenHash).Scan(&dbSession.UserID, &dbSession.ExpiresAt, &ipNull, &dbSession.CreatedAt)
	if err != nil {
		return 0, labelerrors.NewLabelError(constant.LabelRepository+".User.UserIDFromSession.Query", err)
	}
	if ipNull.Valid {
		dbSession.IP = ipNull.String
	}
	dbSession.TokenHash = tokenHash
	_ = r.sessions.Set(ctx, tokenHash, &dbSession)
	return dbSession.UserID, nil
}
func (r *userRepo) IncrementWithdrawn(ctx context.Context, w *model.Withdrawal) error {
	u, err := r.GetByID(ctx, w.UserID)
	if err != nil {
		return err
	}
	u.Balance -= w.Sum
	u.Withdrawn += w.Sum
	_, err = r.repoBase.q(ctx).ExecContext(ctx,
		userIncrementWithdrawnQuery,
		u.Balance, u.Withdrawn, u.ID)
	if err != nil {
		return labelerrors.NewLabelError(constant.LabelRepository+".User.IncrementWithdrawn.Exec", err)
	}
	err = r.usersByID.Set(ctx, u.ID, u)
	if err != nil {
		return labelerrors.NewLabelError(constant.LabelRepository+".User.IncrementWithdrawn.Cache", err)
	}

	return nil
}
func (r *userRepo) IncrementBalance(ctx context.Context, userID int64, increment float64) error {
	if increment == 0 {
		return nil
	}
	u, err := r.GetByID(ctx, userID)
	if err != nil {
		return labelerrors.NewLabelError(constant.LabelRepository+".User.IncrementBalance.GetByID", err)
	}
	fmt.Println(u.ID)
	u.Balance += increment
	fmt.Println(u.Balance)
	_, err = r.repoBase.q(ctx).ExecContext(ctx,
		userIncrementBalanceQuery,
		u.Balance, u.ID)
	if err != nil {
		return labelerrors.NewLabelError(constant.LabelRepository+".User.IncrementBalance.Exec", err)
	}
	err = r.usersByID.Set(ctx, u.ID, u)
	if err != nil {
		return labelerrors.NewLabelError(constant.LabelRepository+".User.IncrementBalance.Cache", err)
	}
	return nil

}
