package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UUID      uuid.UUID
	Username  string
	Password  string
	Email     string
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}
