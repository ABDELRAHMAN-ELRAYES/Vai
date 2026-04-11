package documents

import (
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID           uuid.UUID `json:"id"`
	OwnerID      uuid.UUID `json:"owner_id"`
	Name         string    `json:"name"`
	OriginalName string    `json:"original_name"`
	Size         int64     `json:"size"`
	MimeType     string    `json:"mime_type"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type DocumentChunksContent struct {
	DocumentID  string          `json:"DocumentID"`
	TotalChunks int             `json:"TotalChunks"`
	Chunks      []DocumentChunk `json:"Chunks"`
}

type DocumentChunk struct {
	Text      string `json:"Text"`
	Index     int    `json:"Index"`
	StartChar int    `json:"StartChar"`
	EndChar   int    `json:"EndChar"`
}
type CleanupDraftsJob struct {
	service   *Service
	uploadDir string
	chunksDir string
}