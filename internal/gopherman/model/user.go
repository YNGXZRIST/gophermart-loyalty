package model

import "time"

type User struct {
	ID        int64     `json:"id"`
	Login     string    `json:"login"`
	Pass      string    `json:"-"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
