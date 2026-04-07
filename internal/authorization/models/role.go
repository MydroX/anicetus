package models

import "time"

type Role struct {
	UUID        string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
