package models

import "time"

type Session struct {
	Id             uint64
	UUID           string
	UserId         string
	RefreshToken   string
	LastUsedAt     time.Time
	OS             string
	OSVersion      string
	Browser        string
	BrowserVersion string
	IPv4Address    string
	CreatedAt      time.Time
	ExpiresAt      time.Time
}
