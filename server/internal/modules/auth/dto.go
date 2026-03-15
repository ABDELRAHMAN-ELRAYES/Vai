package auth

import "github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/users"

type RegisterUserPayload struct {
	FirstName string `json:"first_name" validate:"required,max=255"`
	LastName  string `json:"last_name" validate:"required,max=255"`
	Email     string `json:"email" validate:"required,email,max=255"`
	Password  string `json:"password" validate:"required,min=3,max=72"`
}

type AuthenticatePayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

type UserWithTokenResponse struct {
	User  *users.UserResponse `json:"user"`
	Token string              `json:"token"`
}
