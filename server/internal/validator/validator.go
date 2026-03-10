package validator

import "github.com/go-playground/validator/v10"

// Validate is the global validator instance used across the application.
var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}
