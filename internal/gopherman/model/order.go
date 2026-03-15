package model

import "time"

type Order struct {
	UserID    int64     `json:"-"`
	OrderID   string    `json:"order"`
	Status    string    `json:"status"`
	Accrual   *float64  `json:"accrual,omitempty"`
	CreatedAt time.Time `json:"uploaded_at"`
	UpdatedAt time.Time `json:"-"`
}
