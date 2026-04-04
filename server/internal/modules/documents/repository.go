package documents

import (
	"context"
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
			Size:     384,
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
