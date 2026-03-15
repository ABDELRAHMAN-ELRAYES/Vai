package auth

import (
	"context"
	"database/sql"
	"time"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/db"
)

// queryTimeoutDuration is used to bound DB calls.
var queryTimeoutDuration = 5 * time.Second

type Repository struct {
	db db.DBTX
}

func NewRepository(db db.DBTX) *Repository {
	return &Repository{db: db}
}

// Enable the Transaction on the Repository methods without passing tx to each method
func (r *Repository) WithTx(tx *sql.Tx) *Repository {
	return &Repository{db: tx}
}
func (repo *Repository) CreateToken(ctx context.Context, token *Token) error {
	query := `
		INSERT INTO verification_tokens (user_id, token, expired_at)
		VALUES ($1, $2, $3)
	`
	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	_, err := repo.db.ExecContext(
		ctx,
		query,
		token.UserID,
		token.Token,
		token.ExpiredAt,
	)
	return err
}
func (repo *Repository) CleanUpToken(ctx context.Context, userID string) error {
	query := `DELETE FROM verification_tokens WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	_, err := repo.db.ExecContext(ctx, query, userID)
	return err
}
