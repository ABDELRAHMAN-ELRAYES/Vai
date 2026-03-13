package users

type CreateUserPayload struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Email     string `jsono:"email" validate:"required"`
	Password  string `jsono:"password" validate:"required"`
}
