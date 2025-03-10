package dto

type LoginRequest struct {
	Username string `json:"username,omitempty" validate:"omitempty,min=4,max=18"`
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type RefreshTokenRequest struct {
}
