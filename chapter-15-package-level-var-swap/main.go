package package_level_var_swap

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

var randRead = rand.Read

func GenerateID() (string, error) {
	b := make([]byte, 16)
	_, err := randRead(b)
	if err != nil {
		return "", fmt.Errorf("generating random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func GenerateShortCode() string {
	b := make([]byte, 4)
	_, err := randRead(b)
	if err != nil {
		return "fallback"
	}
	return hex.EncodeToString(b)
}
