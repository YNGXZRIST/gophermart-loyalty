package service

import (
	"context"
	"fmt"
	"gophermart-loyalty/internal/gopherman/auth/password"
	"gophermart-loyalty/internal/gopherman/model"
	"net/http"
)

type RegisterInput struct {
	Req model.RegisterRequest
	IP  string
}
type RegisterOutput struct {
	Response
	Token string
}
type LoginInput struct {
	Req model.RegisterRequest
	IP  string
}

type LoginResponse struct {
	Response
	Token string
}

func (s *Service) Register(ctx context.Context, in RegisterInput) RegisterOutput {
	user, err := s.Rep.User.GetByLogin(ctx, in.Req.Login)
	if err != nil {
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
