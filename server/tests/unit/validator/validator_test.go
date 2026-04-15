package validator_test

import (
	"testing"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/validator"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
}

func TestValidator(t *testing.T) {
	t.Run("valid struct", func(t *testing.T) {
		s := TestStruct{
			Name:  "John Doe",
			Email: "john@example.com",
		}
		err := validator.Validate.Struct(s)
		assert.NoError(t, err)
	})

	t.Run("invalid struct - missing name", func(t *testing.T) {
		s := TestStruct{
			Email: "john@example.com",
		}
		err := validator.Validate.Struct(s)
		assert.Error(t, err)
	})

	t.Run("invalid struct - invalid email", func(t *testing.T) {
		s := TestStruct{
			Name:  "John Doe",
			Email: "invalid-email",
		}
		err := validator.Validate.Struct(s)
		assert.Error(t, err)
	})
}
