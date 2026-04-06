package dto

import (
	"time"

	"MydroX/anicetus/internal/common/models"
)

type LoginRequest struct {
	Username string  `json:"username,omitempty" validate:"omitempty,min=4,max=18"`
	Email    string  `json:"email,omitempty"    validate:"omitempty,email"`
	Password string  `json:"password"           validate:"required,min=8,max=72"`
	Session  Session `json:"session"`
}

type SessionLoginRequest struct {
	Browser models.Browser `json:"browser" validate:"required"`
	OS      models.OS      `json:"os"      validate:"required"`
	Device  models.Device  `json:"device"  validate:"required"`
	Time    time.Time      `json:"time"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenRequest struct{}

type RegisterAudienceRequest struct {
	Audience    string         `json:"audience"     validate:"required"`
	ServiceName string         `json:"service_name" validate:"required"`
	Description string         `json:"description"`
	Permissions map[string]any `json:"permissions"`
}

type AssignAudienceRequest struct {
	Audience string `json:"audience" validate:"required"`
}

type AudienceListResponse struct {
	Audiences []string `json:"audiences"`
}

type Session struct {
	IPv4Address    string `json:"ipv4_address"    validate:"required,ipv4"`
	OS             string `json:"os"              validate:"required"`
	OSVersion      string `json:"os_version"      validate:"required"`
	Browser        string `json:"browser"         validate:"required"`
	BrowserVersion string `json:"browser_version" validate:"required"`
}
