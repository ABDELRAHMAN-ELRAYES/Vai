package chat

import (
	"context"
	"database/sql"
	"time"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/db"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/apierror"
)

// queryTimeoutDuration is used to bound DB calls.
var queryTimeoutDuration = 5 * time.Second

type Repository struct {
	db db.DBTX
}

func NewRepository(db db.DBTX) *Repository {
	return &Repository{
		db: db,
	}
}
func (r *Repository) WithTx(tx *sql.Tx) *Repository {
	return &Repository{db: tx}
}
func (repo *Repository) CreateConversation(ctx context.Context, conversation *Conversation) error {
	query := `INSERT INTO conversations(title,user_id)
			  VALUES($1,$2)
			  RETURNING id, created_at`

	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	return repo.db.QueryRowContext(
		ctx,
		query,
		conversation.Title,
		conversation.UserID,
	).Scan(
		&conversation.ID,
		&conversation.CreatedAt,
	)
}
func (repo *Repository) CreateMessage(ctx context.Context, message *Message) error {
	query := `INSERT INTO messages(content,conversation_id,role)
			  VALUES($1,$2,$3)
			  RETURNING id, created_at`

	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	return repo.db.QueryRowContext(
		ctx,
		query,
		message.Content,
		message.ConversationID,
		message.Role,
	).Scan(
		&message.ID,
		&message.CreatedAt,
	)
}

// Update conversation Title
func (repo *Repository) UpdateConversation(ctx context.Context, conv *Conversation) error {
	query := `UPDATE conversations
			  SET title = $1, updated_at = NOW()
			  WHERE id = $2`

	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	result, err := repo.db.ExecContext(ctx, query, conv.ID, conv.Title)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return apierror.ErrNotFound
	}

	return nil
}
