package models

import "time"

type Service struct {
	UUID        string
	Audience    string
	ServiceName string
	Description string
	Permissions map[string]any
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
