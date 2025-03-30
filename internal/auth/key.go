package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"os"
)

var (
	hashSecret = os.Getenv("HASH_SECRET")
)

func GenerateApiKey() (string, error) {
	key := make([]byte, 32, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(key), nil
}

func HashApiKey(key string) string {
	h := hmac.New(sha256.New, []byte(hashSecret))
	h.Write([]byte(key))
	return hex.EncodeToString(h.Sum(nil))
}
