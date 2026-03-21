package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gophermart-loyalty/internal/gopherman/auth/password"
	"gophermart-loyalty/internal/gopherman/model"
	"net/http"
)

// RegisterInput contains payload for user registration.
type RegisterInput struct {
	Req model.RegisterRequest
	IP  string
}

// RegisterOutput contains registration result and session token.
type RegisterOutput struct {
	Response
	Token string
}

// LoginInput contains payload for user authentication.
type LoginInput struct {
	Req model.RegisterRequest
	IP  string
}

// LoginResponse contains authentication result and session token.
type LoginResponse struct {
	Response
	Token string
}

// Register creates a new user and returns a session token.
func (s *Service) Register(ctx context.Context, in RegisterInput) RegisterOutput {
	user, err := s.Rep.User.GetByLogin(ctx, in.Req.Login)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return RegisterOutput{
			Response: Response{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("get user by login failed: %w", err),
			},
		}
	}
	if user != nil {
		return RegisterOutput{
			Response: Response{
				Code: http.StatusConflict,
				Err:  fmt.Errorf("user with login %s already exists", user.Login),
			},
		}
	}
	user, err = s.Rep.User.Register(ctx, in.Req.Login, in.Req.Pass, in.IP)
	if err != nil {
		return RegisterOutput{
			Response: Response{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("register user failed: %w", err),
			},
		}
	}
	session, err := s.Rep.User.CreateSession(ctx, user.ID, in.IP)
	if err != nil {
		return RegisterOutput{
			Response: Response{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("create session failed: %w", err),
			},
		}
	}
	return RegisterOutput{
		Response: Response{Code: http.StatusOK},
		Token:    session,
	}
}

// Login validates credentials and returns a new session token.
func (s *Service) Login(ctx context.Context, in LoginInput) LoginResponse {
	user, err := s.Rep.User.GetByLogin(ctx, in.Req.Login)
	if err != nil {
		return LoginResponse{
			Response: Response{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("get user by login failed: %w", err),
			},
		}
	}
	if user == nil {
		return LoginResponse{
			Response: Response{
				Code: http.StatusUnauthorized,
				Err:  fmt.Errorf("user not found"),
			},
		}
	}
	if err := password.Compare(user.Pass, in.Req.Pass); err != nil {
		return LoginResponse{
			Response: Response{
				Code: http.StatusUnauthorized,
				Err:  fmt.Errorf("invalid password"),
			},
		}
	}
	session, err := s.Rep.User.CreateSession(ctx, user.ID, in.IP)
	if err != nil {
		return LoginResponse{
			Response: Response{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("create session failed: %w", err),
			},
		}
	}
	return LoginResponse{
		Response: Response{Code: http.StatusOK},
		Token:    session,
	}
}
