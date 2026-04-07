package models

import "time"

type Permission struct {
	UUID        string
	Name        string
	Description string
	CreatedAt   time.Time
}
