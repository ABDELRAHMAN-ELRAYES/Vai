package documents

import (
	"context"
	"fmt"
	"time"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/db"
	"github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// queryTimeoutDuration is used to bound DB calls.
var queryTimeoutDuration = 5 * time.Second

type Repository struct {
	db     db.DBTX
	qdrant *QdrantClient
}
type QdrantClient struct {
	client         *qdrant.Client
	collectionName string
}

func NewRepository(db db.DBTX, qdrantClient *QdrantClient) *Repository {

	return &Repository{
		db:     db,
		qdrant: qdrantClient,
	}

}
func (repo *Repository) InitCollection() error {
	ctx := context.Background()

	// Check if the collection already exists before Creating it (return grpc status)
	_, err := repo.qdrant.client.GetCollectionInfo(ctx, repo.qdrant.collectionName)

	if err == nil {
		return nil
	}

	if status.Code(err) != codes.NotFound {
		return err
	}

	err = repo.qdrant.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: repo.qdrant.collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     768,
			Distance: qdrant.Distance_Cosine,
		}),
	})

	return err
}

func (repo *Repository) Create(ctx context.Context, doc *Document) error {
	query := `
		INSERT INTO documents (owner_id, name, original_name, size, mime_type)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, status, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	return repo.db.QueryRowContext(
		ctx,
		query,
		doc.OwnerID,
		doc.Name,
		doc.OriginalName,
		doc.Size,
		doc.MimeType,
	).Scan(
		&doc.ID,
		&doc.Status,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)
}

func (repo *Repository) GetDocumentByID(ctx context.Context, id string) (*Document, error) {
	query := `
		SELECT id, owner_id, name, original_name, size, mime_type, status, created_at, updated_at
		FROM documents
		WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	var doc Document
	err := repo.db.QueryRowContext(ctx, query, id).Scan(
		&doc.ID,
		&doc.OwnerID,
		&doc.Name,
		&doc.OriginalName,
		&doc.Size,
		&doc.MimeType,
		&doc.Status,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (repo *Repository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `
		UPDATE documents
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`
	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	_, err := repo.db.ExecContext(ctx, query, status, id)
	return err
}

func (repo *Repository) DeleteDocument(ctx context.Context, id string) error {
	query := `
		DELETE FROM documents
		WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	_, err := repo.db.ExecContext(ctx, query, id)
	return err
}

func (repo *Repository) UpsertPoints(ctx context.Context, points []*qdrant.PointStruct) error {
	_, err := repo.qdrant.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: repo.qdrant.collectionName,
		Points:         points,
	})
	return err
}

func (repo *Repository) DeletePointsByDocumentID(ctx context.Context, documentID string) error {
	_, err := repo.qdrant.client.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: repo.qdrant.collectionName,
		Points: qdrant.NewPointsSelectorFilter(&qdrant.Filter{
			Must: []*qdrant.Condition{
				qdrant.NewMatchKeyword("document_id", documentID),
			},
		}),
	})
	return err
}

func (repo *Repository) SearchPoints(ctx context.Context, vector []float32, documentIDs []string, topK uint64) ([]*qdrant.ScoredPoint, error) {
	var filter *qdrant.Filter
	if len(documentIDs) > 0 {
		var conditions []*qdrant.Condition
		if len(documentIDs) == 1 {
			conditions = append(conditions, qdrant.NewMatchKeyword("document_id", documentIDs[0]))
		} else {
			conditions = append(conditions, qdrant.NewMatchKeywords("document_id", documentIDs...))
		}
		filter = &qdrant.Filter{
			Must: conditions,
		}
	}

	res, err := repo.qdrant.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: repo.qdrant.collectionName,
		Query:          qdrant.NewQueryNearest(qdrant.NewVectorInputDense(vector)),
		Limit:          qdrant.PtrOf(topK),
		Filter:         filter,
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (repo *Repository) GetOldDrafts(ctx context.Context, olderThan time.Duration) ([]*Document, error) {
	query := `
		SELECT id, owner_id, name, original_name, size, mime_type, status, created_at, updated_at
		FROM documents
		WHERE status = 'draft' AND created_at < NOW() - $1::interval
	`
	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	rows, err := repo.db.QueryContext(ctx, query, fmt.Sprintf("%d seconds", int(olderThan.Seconds())))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*Document
	for rows.Next() {
		var doc Document
		if err := rows.Scan(
			&doc.ID,
			&doc.OwnerID,
			&doc.Name,
			&doc.OriginalName,
			&doc.Size,
			&doc.MimeType,
			&doc.Status,
			&doc.CreatedAt,
			&doc.UpdatedAt,
		); err != nil {
			return nil, err
		}
		docs = append(docs, &doc)
	}
	return docs, nil
}