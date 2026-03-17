package model

import "time"

type Order struct {
	ID        int64     `json:"-"`
	UserID    int64     `json:"-"`
	OrderID   string    `json:"number"`
	Status    string    `json:"status"`
	Accrual   *float64  `json:"accrual,omitempty"`
	CreatedAt time.Time `json:"uploaded_at"`
	UpdatedAt time.Time `json:"-"`
}
