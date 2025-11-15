package model

import "time"

type User struct {
	ID        int        `json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Username  string     `json:"username"`
	Password  string     `json:"password"`
	LastLogin *time.Time `json:"last_login"`
}

