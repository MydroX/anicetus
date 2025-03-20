package dto

import (
	"MydroX/anicetus/internal/common/models"
	"time"
)

type LoginRequest struct {
	Username string  `json:"username,omitempty" validate:"omitempty,min=4,max=18"`
	Email    string  `json:"email,omitempty" validate:"omitempty,email"`
	Password string  `json:"password" validate:"required,min=8,max=72"`
	Session  Session `json:"session"`
}

type SessionLoginRequest struct {
	Browser models.Browser `json:"browser" validate:"required"`
	OS      models.OS      `json:"os" validate:"required"`
	Device  models.Device  `json:"device" validate:"required"`
	Time    time.Time      `json:"time"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenRequest struct {
}

type Session struct {
	RefreshToken   string `json:"refresh_token" validate:"required"`
	IPv4Address    string `json:"ipv4_address" validate:"required,ipv4"`
	OS             string `json:"os" validate:"required"`
	Browser        string `json:"browser" validate:"required"`
	BrowserVersion string `json:"browser_version" validate:"required"`
}
