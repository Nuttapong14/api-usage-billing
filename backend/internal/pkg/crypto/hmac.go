package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// SignSHA256 computes an HMAC-SHA256 signature for the payload.
func SignSHA256(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// SignSHA256WithPrefix returns the signature with an algorithm prefix.
func SignSHA256WithPrefix(secret string, payload []byte) string {
	return "sha256=" + SignSHA256(secret, payload)
}

// VerifySHA256 compares a signature with the payload signature.
func VerifySHA256(secret string, payload []byte, signature string) bool {
	expected := SignSHA256(secret, payload)
	return hmac.Equal([]byte(expected), []byte(signature))
}
