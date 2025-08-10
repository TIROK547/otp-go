package models

import "time"

type User struct {
	ID        int64     `json:"id"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}
