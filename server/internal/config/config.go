package config

import "github.com/ABDELRAHMAN-ELRAYES/Vai/internal/env"

type Config struct {
	Addr        string
	Env         string
	APIURL      string
	FrontendURL string
}

func Load() Config {
	return Config{
		Addr:        env.GetStringEnv("ADDR", ":8080"),
		APIURL:      env.GetStringEnv("API_URL", "localhost:3000"),
		FrontendURL: env.GetStringEnv("FRONTEND_URL", "http://localhost:5173"),
		Env:         env.GetStringEnv("ENV", "development"),
	}
}
