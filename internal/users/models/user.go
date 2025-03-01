package models

import (
	"time"
)

type User struct {
	UUID      string
	Username  string
	Password  string
	Email     string
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}
