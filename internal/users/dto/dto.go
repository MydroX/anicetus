package dto

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=4,max=18"`
	Password string `json:"password" validate:"required,min=14,max=72"`
	Email    string `json:"email" validate:"required,email"`
}

type GetUserResponse struct {
	UUID     string   `json:"uuid"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Role     []string `json:"role"`
}

type GetAllUsersResponse struct {
	Users []*User `json:"users"`
}

type User struct {
	UUID     string   `json:"uuid"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Role     []string `json:"role"`
}

type UpdateUserRequest struct {
	UUID     string   `json:"uuid" validate:"required,min=36,max=46"`
	Username string   `json:"username" validate:"required,min=4,max=18"`
	Password string   `json:"password" validate:"required,min=14,max=72"`
	Email    string   `json:"email" validate:"required,email"`
	Role     []string `json:"role" validate:"required"`
}

type UpdatePasswordRequest struct {
	Password string `json:"password" validate:"required,min=14,max=72"`
}

type UpdateEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}
