package config

import (
	"os"
	"strconv"
)

// positiveInt reads a positive integer environment value or uses its fallback.
func positiveInt(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
