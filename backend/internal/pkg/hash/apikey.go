package hash

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

const (
	DefaultLivePrefix = "sk_live_"
	DefaultTestPrefix = "sk_test_"
	DefaultPrefixLen  = 12
	defaultKeyBytes   = 32
)

// HashAPIKey returns the SHA-256 hash of an API key.
func HashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// GenerateAPIKey creates a new API key with the provided prefix.
func GenerateAPIKey(prefix string) (string, string, error) {
	if prefix == "" {
		prefix = DefaultLivePrefix
	}

	randomBytes := make([]byte, defaultKeyBytes)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", err
	}

	payload := base64.RawURLEncoding.EncodeToString(randomBytes)
	plain := prefix + payload
	return plain, HashAPIKey(plain), nil
}

// APIKeyPrefix returns the display prefix for an API key.
func APIKeyPrefix(key string) string {
	if len(key) <= DefaultPrefixLen {
		return key
	}
	return key[:DefaultPrefixLen]
}
