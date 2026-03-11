package config

import "github.com/ABDELRAHMAN-ELRAYES/Vai/internal/env"

type Config struct {
	Addr        string
	Env         string
	APIURL      string
	FrontendURL string
	DB          DB
	AI          AI
	QdrantDB    QdrantConfig
}

// DB holds database related configuration.
type DB struct {
	Addr         string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}
type AI struct {
	BaseURL        string
	Name           string
	EmBeddingModel string
}
type QdrantConfig struct {
	Host string
	Port int
}

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
		AI: AI{
			BaseURL:        env.GetStringEnv("AI_MODEL_URL", "http://localhost:11434"),
			Name:           env.GetStringEnv("AI_MODEL_NAME", "qwen3.5:4b"),
			EmBeddingModel: env.GetStringEnv("AI_MODEL_EMBEDDING_NAME", "nomic-embed-text:v1.5"),
		},
		QdrantDB: QdrantConfig{
			Host: env.GetStringEnv("QDRANT_DB_HOST", "localhost"),
			Port: env.GetIntEnv("QDRANT_DB_PORT", 6334),
		},
	}
}
