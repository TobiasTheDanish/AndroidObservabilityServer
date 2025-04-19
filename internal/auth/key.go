package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
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

func HashPassword(pw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)

	return string(hash), err
}

func ValidatePassword(pw, pwHash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(pwHash), []byte(pw)) == nil
}

func GenerateSessionToken() (string, error) {
	key := make([]byte, 32, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(key), nil
}

func GetExpiryForSession() int64 {
	return time.Now().UTC().AddDate(0, 0, 7).UnixMilli()
}
