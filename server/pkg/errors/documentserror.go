package apierror

import "errors"

var (
	ErrEmbedChunksFailed     = errors.New("failed to embed document chunks")
	ErrReadChunksFailed      = errors.New("failed to read chunks file")
	ErrUnmarshalChunksFailed = errors.New("failed to unmarshal chunks file")
	ErrUpsertVectorsFailed   = errors.New("failed to upsert vectors into qdrant")
)
