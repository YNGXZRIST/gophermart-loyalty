package model

import "time"

type Order struct {
	OrderID   string    `json:"order"`
	Status    string    `json:"status"`
	Accrual   *int      `json:"accrual,omitempty"`
	CreatedAt time.Time `json:"uploaded_at"`
	UpdatedAt time.Time `json:"-"`
}
