package chat

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/db"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/documents"
	apierror "github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/errors"
	"github.com/google/uuid"
)

// queryTimeoutDuration is used to bound DB calls.
var queryTimeoutDuration = 5 * time.Second

type IRepository interface {
	CreateConversation(ctx context.Context, conversation *Conversation) error
	CreateMessage(ctx context.Context, message *Message) error
	UpdateConversation(ctx context.Context, conv *Conversation) error
	GetConversationsByUserID(ctx context.Context, userID string) ([]*Conversation, error)
	DeleteConversation(ctx context.Context, id string) error
	GetConversationByID(ctx context.Context, id string) (*Conversation, error)
	GetMessagesByConversationID(ctx context.Context, conversationID string) ([]Message, error)
	AddMessageDocuments(ctx context.Context, messageID string, documentIDs []string) error
	GetAssociatedDocumentIDs(ctx context.Context, conversationID string) ([]string, error)
	WithTx(tx *sql.Tx) IRepository
}
type Repository struct {
	db db.DBTX
}

func NewRepository(db db.DBTX) *Repository {
	return &Repository{
		db: db,
	}
}
func (r *Repository) WithTx(tx *sql.Tx) IRepository {
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

func (repo *Repository) GetConversationByID(ctx context.Context, id string) (*Conversation, error) {
	query := `SELECT id, user_id, title, created_at, updated_at
			  FROM conversations
			  WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	conv := &Conversation{}
	err := repo.db.QueryRowContext(ctx, query, id).Scan(
		&conv.ID,
		&conv.UserID,
		&conv.Title,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apierror.ErrNotFound
		}
		return nil, err
	}

	return conv, nil
}
func (repo *Repository) GetMessagesByConversationID(ctx context.Context, conversationID string) ([]Message, error) {
	query := `
		SELECT 
			m.id, m.conversation_id, m.content, m.role, m.created_at,
			d.id, d.owner_id, d.name, d.original_name, d.size, d.mime_type, d.status, d.created_at, d.updated_at
		FROM messages m 
		LEFT JOIN message_documents md ON m.id = md.message_id
		LEFT JOIN documents d ON md.document_id = d.id
		WHERE m.conversation_id = $1
		ORDER BY m.created_at ASC`

	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	rows, err := repo.db.QueryContext(ctx, query, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messagesMap := make(map[string]*Message)
	var messageOrder []string

	// Scan + Filter duplicates
	for rows.Next() {
		var (
			msg               Message
			doc               documents.Document
			documentID        sql.NullString
			documentOwnerID   sql.NullString
			documentName      sql.NullString
			documentOrigName  sql.NullString
			documentSize      sql.NullInt64
			documentMimeType  sql.NullString
			documentStatus    sql.NullString
			documentCreatedAt sql.NullTime
			documentUpdatedAt sql.NullTime
		)

		err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.Content,
			&msg.Role,
			&msg.CreatedAt,
			&documentID,
			&documentOwnerID,
			&documentName,
			&documentOrigName,
			&documentSize,
			&documentMimeType,
			&documentStatus,
			&documentCreatedAt,
			&documentUpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if _, exists := messagesMap[msg.ID]; !exists {
			msg.Documents = []documents.Document{}
			messagesMap[msg.ID] = &msg
			messageOrder = append(messageOrder, msg.ID)
		}

		if documentID.Valid {
			id, _ := uuid.Parse(documentID.String)
			ownerID, _ := uuid.Parse(documentOwnerID.String)
			doc = documents.Document{
				ID:           id,
				OwnerID:      ownerID,
				Name:         documentName.String,
				OriginalName: documentOrigName.String,
				Size:         documentSize.Int64,
				MimeType:     documentMimeType.String,
				Status:       documentStatus.String,
				CreatedAt:    documentCreatedAt.Time,
				UpdatedAt:    documentUpdatedAt.Time,
			}
			messagesMap[msg.ID].Documents = append(messagesMap[msg.ID].Documents, doc)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	messages := make([]Message, 0, len(messageOrder))
	for _, id := range messageOrder {
		messages = append(messages, *messagesMap[id])
	}

	return messages, nil
}
func (repo *Repository) AddMessageDocuments(ctx context.Context, messageID string, documentIDs []string) error {
	if len(documentIDs) == 0 {
		return nil
	}

	query := `INSERT INTO message_documents(message_id, document_id) VALUES `
	values := []any{}
	for i, docID := range documentIDs {
		if i > 0 {
			query += ","
		}
		query += fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2)
		values = append(values, messageID, docID)
	}

	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	_, err := repo.db.ExecContext(ctx, query, values...)
	return err
}

func (repo *Repository) GetAssociatedDocumentIDs(ctx context.Context, conversationID string) ([]string, error) {
	query := `
		SELECT DISTINCT md.document_id
		FROM message_documents md
		JOIN messages m ON md.message_id = m.id
		WHERE m.conversation_id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	rows, err := repo.db.QueryContext(ctx, query, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docIDs []string
	for rows.Next() {
		var docID string
		if err := rows.Scan(&docID); err != nil {
			return nil, err
		}
		docIDs = append(docIDs, docID)
	}

	return docIDs, rows.Err()
}
