package model

import "time"

type Withdrawal struct {
	UserID    int64     `json:"-"`
	OrderID   string    `json:"order"`
	Sum       float64   `json:"sum"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"processed_at"`
}
