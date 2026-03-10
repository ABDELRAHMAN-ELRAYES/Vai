package users

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByID(id string) (*User, error) {

	row := r.db.QueryRow(
		"SELECT id, name FROM users WHERE id=$1",
		id,
	)

	var user User

	err := row.Scan(&user.ID, &user.Name)
	if err != nil {
		return nil, err
	}

	return &user, nil
}