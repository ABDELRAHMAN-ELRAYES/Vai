package users

import "golang.org/x/crypto/bcrypt"

type User struct {
	ID        string   `json:"id"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Email     string   `json:"email"`
	Password  Password `json:"-"` // The password musn't be returned even on the hashed state
	CreatedAt string   `json:"created_at"`
	IsActive  bool     `json:"is_active"`
}

type Password struct {
	Text *string
	Hash []byte
}

// Hash & Compare the entered plain text password
func (password *Password) Set(text string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	password.Text = &text
	password.Hash = hashed
	return nil
}

func (password *Password) Compare(text string) error {
	return bcrypt.CompareHashAndPassword(password.Hash, []byte(text))
}
