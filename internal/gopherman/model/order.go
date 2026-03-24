package model

import "time"

// Order represents loyalty order entity and API payload fields.
type Order struct {
	ID        int64     `json:"-"`
	UserID    int64     `json:"-"`
	OrderID   string    `json:"number"`
	Status    string    `json:"status"`
	Accrual   *float64  `json:"accrual,omitempty"`
	CreatedAt time.Time `json:"uploaded_at"`
	UpdatedAt time.Time `json:"-"`
}
