package model

import "time"

type Sessions struct {
	ID        int64
	UserID    int64
	TokenHash string
	IP        string
	ExpiresAt time.Time
	CreatedAt time.Time
}
