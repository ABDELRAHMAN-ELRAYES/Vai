package users

import (
	"context"
	"database/sql"
	"time"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/apierror"
	"github.com/google/uuid"
)

// queryTimeoutDuration is used to bound DB calls.
var queryTimeoutDuration = 5 * time.Second

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}
func (repo *Repository) Create(ctx context.Context, user *User) error {

	query := `
		INSERT INTO users(first_name, last_name, email, password)
		VALUES($1, $2, $3, $4) 
		RETURNING id, created_at
	`
	// Set Timeout on the context for limiting the query execution duration
	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	return repo.db.QueryRowContext(
		ctx,
		query,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Password.Hash,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)
}
func (repo *Repository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {

	query := `SELECT id, first_name, last_name, email, created_at, is_active FROM users WHERE id=$1`

	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	user := &User{}

	err := repo.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive,
	)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, apierror.ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}
