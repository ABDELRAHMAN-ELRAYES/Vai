package env

import (
	"os"
	"strconv"
)

func GetStringEnv(key, fallback string) string {
	val, ok := os.LookupEnv(key)

	if !ok {
		return fallback
	}
	return val
}

func GetIntEnv(key string, fallback int) int {
	val, ok := os.LookupEnv(key)

	if !ok {
		return fallback
	}
	valAsInt, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return valAsInt
}

func GetBoolEnv(key string, fallback bool) bool {
	val, ok := os.LookupEnv(key)

	if !ok {
		return fallback
	}
	valAsBool, err := strconv.ParseBool(val)
	if err != nil {
		return fallback
	}
	return valAsBool
}
