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
// Update conversation Title
func (repo *Repository) UpdateConversation(ctx context.Context, conv *Conversation) error {
	query := `UPDATE conversations
			  SET title = $1, updated_at = NOW()
			  WHERE id = $2`

	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	result, err := repo.db.ExecContext(ctx, query, conv.Title, conv.ID)
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

func (repo *Repository) GetConversationsByUserID(ctx context.Context, userID string) ([]*Conversation, error) {
	query := `SELECT id, user_id, title, created_at, updated_at
			  FROM conversations
			  WHERE user_id = $1
			  ORDER BY updated_at DESC`

	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	rows, err := repo.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []*Conversation
	for rows.Next() {
		conv := &Conversation{}
		err := rows.Scan(
			&conv.ID,
			&conv.UserID,
			&conv.Title,
			&conv.CreatedAt,
			&conv.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, conv)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return conversations, nil
}

func (repo *Repository) DeleteConversation(ctx context.Context, id string) error {
	query := `DELETE FROM conversations WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	result, err := repo.db.ExecContext(ctx, query, id)
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
