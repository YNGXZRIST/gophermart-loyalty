package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"gophermart-loyalty/internal/gopherman/auth/password"
	"gophermart-loyalty/internal/gopherman/auth/session"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/errors/labelerrors"
	"gophermart-loyalty/internal/gopherman/model"
	"strconv"
	"time"

	"github.com/patrickmn/go-cache"
)

type UserRepo struct {
	repoBase  repoBase
	loginToID *cache.Cache
	usersByID *cache.Cache
	sessions  *cache.Cache
	//loginToID *storage.MemStorage[string, int64]
	//usersByID *storage.MemStorage[int64, *model.User]
	//sessions  *storage.MemStorage[string, *model.Sessions]
}

const (
	UserGetByLoginQuery    = "SELECT id, login, pass, created_at, updated_at, last_login_ip, balance, withdrawn FROM users WHERE login=$1"
	UserGetByIDQuery       = "SELECT id, login, pass, created_at, updated_at, last_login_ip, balance, withdrawn FROM users WHERE id=$1"
	UserRegisterQuery      = "INSERT INTO users(login, pass, last_login_ip) VALUES ($1, $2, $3) RETURNING id, login, pass, created_at, updated_at, last_login_ip"
	UserUpdateLastIPQuery  = "UPDATE users SET last_login_ip=$1, updated_at = CURRENT_TIMESTAMP where id=$2"
	UserUpsertSessionQuery = `INSERT INTO sessions (token_hash, user_id, expires_at, ip) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (user_id, ip) DO UPDATE SET
		   token_hash = EXCLUDED.token_hash,
		   expires_at = EXCLUDED.expires_at`
	UserUserIDFromSessionQuery = "SELECT user_id, expires_at, ip, created_at FROM sessions WHERE token_hash = $1 AND expires_at > CURRENT_TIMESTAMP"

	UserIncrementWithdrawnQuery = `UPDATE users SET balance = $1,withdrawn = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3`
	UserIncrementBalanceQuery   = "UPDATE users SET balance = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2;"
)

func NewUserRepository(db *conn.DB) *UserRepo {
	loginToID := cache.New(5*time.Minute, 10*time.Minute)
	usersByID := cache.New(5*time.Minute, 10*time.Minute)
	sessStorage := cache.New(5*time.Minute, 10*time.Minute)

	return &UserRepo{repoBase: repoBase{db: db}, loginToID: loginToID, usersByID: usersByID, sessions: sessStorage}
}

func (r *UserRepo) setUserCache(u *model.User) {
	r.loginToID.Set(u.Login, u.ID, cache.DefaultExpiration)
	r.usersByID.Set(strconv.FormatInt(u.ID, 10), u, cache.DefaultExpiration)
}

func (r *UserRepo) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	id, found := r.loginToID.Get(login)
	if found {
		return r.GetByID(ctx, id.(int64))
	}
	u, err := r.GetByLoginForce(ctx, login)
	if err != nil {
		return nil, err
	}
	r.setUserCache(u)
	return u, nil
}

func (r *UserRepo) GetByLoginForce(ctx context.Context, login string) (*model.User, error) {
	var dbUser model.User
	var lastIP sql.NullString
	err := r.repoBase.q(ctx).QueryRowContext(ctx,
		UserGetByLoginQuery,
		login,
	).Scan(&dbUser.ID, &dbUser.Login, &dbUser.Pass, &dbUser.CreatedAt, &dbUser.UpdatedAt, &lastIP, &dbUser.Balance, &dbUser.Withdrawn)
	if err != nil {
		return nil, labelerrors.NewLabelError(labelRepository+".User.GetByLogin.Query", err)
	}
	if lastIP.Valid {
		dbUser.LastIP = lastIP.String
	}
	return &dbUser, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	c, found := r.usersByID.Get(strconv.FormatInt(id, 10))
	if found {
		u, ok := c.(*model.User)
		if ok {
			return u, nil
		}
	}
	u, err := r.GetByIDForce(ctx, id)
	if err != nil {
		return nil, err
	}
	r.setUserCache(u)
	return u, nil
}
func (r *UserRepo) GetByIDForce(ctx context.Context, id int64) (*model.User, error) {
	var dbUser model.User
	var lastIP sql.NullString
	err := r.repoBase.q(ctx).QueryRowContext(ctx,
		UserGetByIDQuery,
		id,
	).Scan(&dbUser.ID, &dbUser.Login, &dbUser.Pass, &dbUser.CreatedAt, &dbUser.UpdatedAt, &lastIP, &dbUser.Balance, &dbUser.Withdrawn)
	if err != nil {
		return nil, labelerrors.NewLabelError(labelRepository+".User.GetByID.Query", err)
	}
	if lastIP.Valid {
		dbUser.LastIP = lastIP.String
	}
	return &dbUser, nil
}

func (r *UserRepo) Register(ctx context.Context, login, pass, ip string) (*model.User, error) {
	hash, err := password.Hash(pass)
	if err != nil {
		return nil, labelerrors.NewLabelError(labelRepository+".User.Register.HashPassword", err)
	}
	var u model.User
	var lastIP sql.NullString
	err = r.repoBase.q(ctx).QueryRowContext(ctx,
		UserRegisterQuery,
		login, hash, ip,
	).Scan(&u.ID, &u.Login, &u.Pass, &u.CreatedAt, &u.UpdatedAt, &lastIP)
	if err != nil {
		return nil, labelerrors.NewLabelError(labelRepository+".User.Register.Insert", err)
	}
	if lastIP.Valid {
		u.LastIP = lastIP.String
	}
	r.setUserCache(&u)
	return &u, nil
}
func (r *UserRepo) UpdateLastIP(ctx context.Context, userID int64, ip string) error {
	u, err := r.GetByIDForce(ctx, userID)
	if err != nil {
		return labelerrors.NewLabelError(labelRepository+".User.UpdateLastIP.GetByID", err)
	}
	_, err = r.repoBase.q(ctx).ExecContext(ctx, UserUpdateLastIPQuery, ip, userID)
	if err != nil {
		return labelerrors.NewLabelError(labelRepository+".User.UpdateLastIP.Exec", err)
	}
	u.LastIP = ip
	r.setUserCache(u)
	return nil
}
func (r *UserRepo) CreateSession(ctx context.Context, userID int64, ip string) (string, error) {
	token, err := session.GenerateToken()
	if err != nil {
		return "", labelerrors.NewLabelError(labelRepository+".User.CreateSession.GenerateToken", err)
	}
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])
	expiresAt := time.Now().Add(24 * time.Hour)
	err = r.UpdateLastIP(ctx, userID, ip)
	if err != nil {
		return "", err
	}
	_, err = r.repoBase.q(ctx).ExecContext(ctx,
		UserUpsertSessionQuery,
		tokenHash, userID, expiresAt, ip)
	if err != nil {
		return "", labelerrors.NewLabelError(labelRepository+".User.CreateSession.Upsert", err)
	}
	r.sessions.Set(tokenHash, &model.Sessions{UserID: userID, TokenHash: tokenHash, ExpiresAt: expiresAt, IP: ip}, cache.DefaultExpiration)
	return token, nil
}

func (r *UserRepo) UserIDFromSession(ctx context.Context, token string) (int64, error) {
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])
	c, found := r.sessions.Get(tokenHash)
	if found {
		s, ok := c.(*model.Sessions)
		if ok {
			if !s.ExpiresAt.After(time.Now()) {
				return 0, nil
			}
			return s.UserID, nil
		}
	}
	var dbSession model.Sessions
	var ipNull sql.NullString
	err := r.repoBase.q(ctx).QueryRowContext(ctx,
		UserUserIDFromSessionQuery,
		tokenHash).Scan(&dbSession.UserID, &dbSession.ExpiresAt, &ipNull, &dbSession.CreatedAt)
	if err != nil {
		return 0, labelerrors.NewLabelError(labelRepository+".User.UserIDFromSession.Query", err)
	}
	if ipNull.Valid {
		dbSession.IP = ipNull.String
	}
	dbSession.TokenHash = tokenHash
	r.sessions.Set(tokenHash, &dbSession, cache.DefaultExpiration)
	return dbSession.UserID, nil
}
func (r *UserRepo) IncrementWithdrawn(ctx context.Context, w *model.Withdrawal) error {
	u, err := r.GetByIDForce(ctx, w.UserID)
	if err != nil {
		return err
	}
	newBalance := u.Balance - w.Sum
	newWithdrawn := u.Withdrawn + w.Sum
	_, err = r.repoBase.q(ctx).ExecContext(ctx,
		UserIncrementWithdrawnQuery,
		newBalance, newWithdrawn, u.ID)
	if err != nil {
		return labelerrors.NewLabelError(labelRepository+".User.IncrementWithdrawn.Exec", err)
	}
	u.Balance = newBalance
	u.Withdrawn = newWithdrawn
	r.setUserCache(u)
	return nil
}
func (r *UserRepo) IncrementBalance(ctx context.Context, userID int64, increment float64) error {
	if increment == 0 {
		return nil
	}
	u, err := r.GetByIDForce(ctx, userID)
	if err != nil {
		return labelerrors.NewLabelError(labelRepository+".User.IncrementBalance.GetByID", err)
	}
	newBalance := u.Balance + increment
	_, err = r.repoBase.q(ctx).ExecContext(ctx,
		UserIncrementBalanceQuery,
		newBalance, u.ID)
	if err != nil {
		return labelerrors.NewLabelError(labelRepository+".User.IncrementBalance.Exec", err)
	}
	u.Balance = newBalance
	r.setUserCache(u)

	return nil

}
