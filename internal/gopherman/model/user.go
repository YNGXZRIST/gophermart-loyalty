package model

import (
	"errors"
	"time"
)

type User struct {
	ID        int64
	Login     string
	Pass      string
	LastIp    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RegisterRequest struct {
	Login string `json:"login"`
	Pass  string `json:"password"`
}

func (r *RegisterRequest) Validate() error {
	if len(r.Login) < 3 || len(r.Login) > 255 {
		return errors.New("login must be between 3 and 255 chars")
	}
	if len(r.Pass) < 5 {
		return errors.New("password too short")
	}
	return nil
}

func (r *RegisterRequest) ToUser() User {
	return User{Login: r.Login, Pass: r.Pass}
}
