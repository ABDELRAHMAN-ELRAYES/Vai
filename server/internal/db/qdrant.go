package db

import (
	"github.com/qdrant/go-client/qdrant"
)

// Create a qdrant client
func NewQdrantClient(host string, port int) (*qdrant.Client, error) {

	client, err := qdrant.NewClient(&qdrant.Config{
		Host: host,
		Port: port,
	})

	if err != nil {
		return nil, err
	}

	return client, nil
}
