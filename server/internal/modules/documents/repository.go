package documents

import (
	"context"
	"database/sql"

	"github.com/qdrant/go-client/qdrant"
)

type Repository struct {
	db     *sql.DB
	qdrant *QdrantClient
}
type QdrantClient struct {
	client         *qdrant.Client
	collectionName string
}

func NewRepository(db *sql.DB, qdrantClient *QdrantClient) *Repository {

	return &Repository{
		db:     db,
		qdrant: qdrantClient,
	}

}
func (repo *Repository) InitCollection() error {
	ctx := context.Background()

	// Check if the collection already exists before Creating it
	exists, err := repo.qdrant.client.GetCollectionInfo(ctx, repo.qdrant.collectionName)
	if err != nil {
		return err
	}

	if exists != nil {
		return nil
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
