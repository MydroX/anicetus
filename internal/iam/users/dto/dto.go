package dto

import "github.com/google/uuid"

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=4,max=18"`
	Password string `json:"password" validate:"required,min=14,max=72"`
	Email    string `json:"email" validate:"required,email"`
	Role     string `json:"role" validate:"required,oneof=ADMIN USER"`
}

type GetUserResponse struct {
	UUID     uuid.UUID `json:"uuid"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
}

type UpdateUserRequest struct {
	UUID     uuid.UUID `json:"uuid"`
	Username string    `json:"username" validate:"required,min=4,max=18"`
	Password string    `json:"password" validate:"required,min=14,max=72"`
	Email    string    `json:"email" validate:"required,email"`
	Role     string    `json:"role" validate:"required,oneof=ADMIN USER"`
}

type UpdatePasswordRequest struct {
	Password string `json:"password" validate:"required,min=14,max=72"`
}

type UpdateEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}
