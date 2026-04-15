package validator

import v "github.com/go-playground/validator/v10"

// Validate is the global validator instance used across the application.
var Validate *v.Validate

func init() {
	Validate = v.New(v.WithRequiredStructEnabled())
}
