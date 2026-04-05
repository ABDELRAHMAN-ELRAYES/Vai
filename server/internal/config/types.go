package config

import "time"

type Config struct {
	Addr          string
	Env           string
	APIURL        string
	FrontendURL   string
	DB            DB
	RAG           RAGConfig
	QdrantDB      QdrantConfig
	Authenticator AuthenticatorConfig
	Mail          Mail
	Upload        UploadConfig
}

// DB holds database related configuration.
type DB struct {
	Addr         string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}
type RAGConfig struct {
	AI      AI
	Chunker ChunkerConfig
}
type AI struct {
	BaseURL        string
	Name           string
	EmBeddingModel string
}
type ChunkerConfig struct {
	ChunkSize int
	Overlap   int
	ChunksDir string
}
type QdrantConfig struct {
	Host string
	Port int
}
type AuthenticatorConfig struct {
	JWT JWTConfig
}
type JWTConfig struct {
	Secret       string
	Iss          string
	Aud          string
	SessionExp   time.Duration
	MailTokenExp time.Duration
}
type Mail struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	FromName     string
	FromAddress  string
	SupportEmail string
	Expiry       time.Duration
}
type UploadConfig struct {
	Dir string
}
