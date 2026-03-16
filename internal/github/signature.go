package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func VerifySignature(payload []byte, signature string, secret string) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	expectedSignature := strings.TrimPrefix(signature, "sha256=")
	expectedDigest, err := hex.DecodeString(expectedSignature)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	computedDigest := mac.Sum(nil)

	return hmac.Equal(computedDigest, expectedDigest)
}
