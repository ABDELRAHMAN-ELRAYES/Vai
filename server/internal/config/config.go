package config

import (
	"time"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/env"
)

func Load() Config {
	return Config{
		Addr:        env.GetStringEnv("ADDR", ":8080"),
		APIURL:      env.GetStringEnv("API_URL", "localhost:3000"),
		FrontendURL: env.GetStringEnv("FRONTEND_URL", "http://localhost:5173"),
		Env:         env.GetStringEnv("ENV", "development"),
		DB: DB{
			Addr:         env.GetStringEnv("DB_ADDR", "postgres://postgres:password@localhost/social_db?sslmode=disable"),
			MaxOpenConns: env.GetIntEnv("DB_MAX_OPEN_CONNS", 30),
			MaxIdleConns: env.GetIntEnv("DB_MAX_IDLE_CONNS", 30),
			MaxIdleTime:  env.GetStringEnv("DB_MAX_IDLE_TIME", "15m"),
		},
		RAG: RAGConfig{
			AI: AI{
				BaseURL:        env.GetStringEnv("RAG_AI_MODEL_URL", "http://localhost:11434"),
				Name:           env.GetStringEnv("RAG_AI_MODEL_NAME", "llama3.2:3b"),
				EmBeddingModel: env.GetStringEnv("RAG_AI_MODEL_EMBEDDING_NAME", "nomic-embed-text:v1.5"),
			},
			Chunker: ChunkerConfig{
				ChunkSize: env.GetIntEnv("RAG_CHUNKER_CHUNK_SIZE", 512),
				Overlap:   env.GetIntEnv("RAG_CHUNKER_OVERLAP_SIZE", 70),
				ChunksDir: env.GetStringEnv("UPLOAD_CHUNKS_DIR", "./uploads/chunks"),
			},
		},
		QdrantDB: QdrantConfig{
			Host: env.GetStringEnv("QDRANT_DB_HOST", "localhost"),
			Port: env.GetIntEnv("QDRANT_DB_PORT", 6334),
		},
		Authenticator: AuthenticatorConfig{
			JWT: JWTConfig{
				Secret:       env.GetStringEnv("AUTH_JWT_SECRET", ""),
				Iss:          env.GetStringEnv("AUTH_JWT_ISSUER", "vai-server"),
				Aud:          env.GetStringEnv("AUTH_JWT_AUDIENCE", "users"),
				SessionExp:   90 * 24 * time.Hour,
				MailTokenExp: 15 * time.Minute,
			},
		},
		Mail: Mail{
			SMTPHost:     env.GetStringEnv("MAIL_SMTP_HOST", ""),
			SMTPPort:     env.GetIntEnv("MAIL_SMTP_PORT", 0),
			SMTPUser:     env.GetStringEnv("MAIL_USER", ""),
			SMTPPassword: env.GetStringEnv("MAIL_PASSWORD", ""),
			FromName:     env.GetStringEnv("MAIL_FROM_NAME", ""),
			FromAddress:  env.GetStringEnv("FROM_ADDRESS", ""),
			SupportEmail: env.GetStringEnv("MAIL_SUPPORT_EMAIL", ""),
			Expiry:       73 * time.Hour,
		},
		Upload: UploadConfig{
			Dir: env.GetStringEnv("UPLOAD_DIR", "./uploads/raw"),
		},
	}
}
