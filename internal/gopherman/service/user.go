package service

import (
	"context"
	"fmt"
	"gophermart-loyalty/internal/gopherman/auth/password"
	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/internal/gopherman/repository"
	"net/http"
)

type RegisterResponse struct {
	Response
	Token string
}
type RegisterInput struct {
	Req model.RegisterRequest
	IP  string
}

type LoginInput struct {
	Req model.RegisterRequest
	IP  string
}

type LoginResponse struct {
	Response
	Token string
}

func Register(ctx context.Context, repo repository.UserRepository, in RegisterInput) RegisterResponse {
	user, err := repo.GetByLogin(ctx, in.Req.Login)
	if err != nil {
		return RegisterResponse{
			Response: Response{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("get user by login failed: %w", err),
			},
		}
	}
	if user != nil {
		return RegisterResponse{
			Response: Response{
				Code: http.StatusConflict,
				Err:  fmt.Errorf("user with login %s already exists", user.Login),
			},
		}
	}
	user, err = repo.Register(ctx, in.Req.Login, in.Req.Pass, in.IP)
	if err != nil {
		return RegisterResponse{
			Response: Response{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("register user failed: %w", err),
			},
		}
	}
	session, err := repo.CreateSession(ctx, user.ID, in.IP)
	if err != nil {
		return RegisterResponse{
			Response: Response{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("create session failed: %w", err),
			},
		}
	}
	return RegisterResponse{
		Response: Response{Code: http.StatusOK},
		Token:    session,
	}
}

func Login(ctx context.Context, repo repository.UserRepository, in LoginInput) LoginResponse {
	user, err := repo.GetByLogin(ctx, in.Req.Login)
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
	session, err := repo.CreateSession(ctx, user.ID, in.IP)
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
