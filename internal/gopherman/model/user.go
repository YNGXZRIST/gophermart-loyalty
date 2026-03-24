// Package model defines domain entities and transport payloads.
package model

import (
	"errors"
	"time"
)

// User represents an account in storage layer.
type User struct {
	ID        int64
	Login     string
	Pass      string
	LastIP    string
	Balance   float64
	Withdrawn float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// RegisterRequest is user registration/login payload.
type RegisterRequest struct {
	Login string `json:"login"`
	Pass  string `json:"password"`
}

// BalanceResponse is API payload with current and withdrawn totals.
type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// Validate performs basic request fields validation.
func (r *RegisterRequest) Validate() error {
	if len(r.Login) < 3 || len(r.Login) > 255 {
		return errors.New("login must be between 3 and 255 chars")
	}
	if len(r.Pass) < 5 {
		return errors.New("password too short")
	}
	return nil
}
