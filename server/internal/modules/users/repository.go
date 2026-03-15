package users

import (
	"context"
	"database/sql"
	"time"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/db"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/apierror"
	"github.com/google/uuid"
)

// queryTimeoutDuration is used to bound DB calls.
var queryTimeoutDuration = 5 * time.Second

type Repository struct {
	db db.DBTX
}

func NewRepository(db db.DBTX) *Repository {
	return &Repository{db: db}
}
func (r *Repository) WithTx(tx *sql.Tx) *Repository {
	return &Repository{db: tx}
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
func (repo *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {

	query := `SELECT id, first_name, last_name, email, created_at, is_active FROM users WHERE email=$1`

	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	user := &User{}

	err := repo.db.QueryRowContext(
		ctx,
		query,
		email,
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
func (repo *Repository) ActivateUser(ctx context.Context, user *User) error {
	query := `UPDATE users SET is_active = $1 WHERE id = $2`

	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	_, err := repo.db.ExecContext(ctx, query, user.IsActive, user.ID)
	return err
}

func (repo *Repository) GetFromToken(ctx context.Context, token []byte) (*User, error) {
	query := `
	SELECT users.id, users.first_name, users.last_name, users.email, users.created_at, users.is_active 
	FROM verification_tokens vt JOIN users ON vt.user_id = users.id 
	WHERE vt.token = $1 AND vt.expired_at > $2
	`

	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := repo.db.QueryRowContext(ctx, query, token, time.Now()).Scan(&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive)
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
