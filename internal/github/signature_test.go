package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestVerifySignatureAcceptsValidSignature(t *testing.T) {
	payload := []byte(`{"zen":"keep it logically awesome"}`)
	secret := "super-secret"
	signature := signPayload(payload, secret)

	if !VerifySignature(payload, signature, secret) {
		t.Fatal("expected valid signature to pass verification")
	}
}

func TestVerifySignatureRejectsInvalidSignature(t *testing.T) {
	payload := []byte(`{"zen":"keep it logically awesome"}`)
	secret := "super-secret"

	if VerifySignature(payload, "sha256=deadbeef", secret) {
		t.Fatal("expected invalid signature to fail verification")
	}
}

func TestVerifySignatureRejectsMalformedHeader(t *testing.T) {
	payload := []byte(`{"zen":"keep it logically awesome"}`)
	secret := "super-secret"

	if VerifySignature(payload, "deadbeef", secret) {
		t.Fatal("expected malformed signature header to fail verification")
	}
}

func TestVerifySignatureAllowsUppercaseHex(t *testing.T) {
	payload := []byte(`{"zen":"keep it logically awesome"}`)
	secret := "super-secret"
	signature := "sha256=" + strings.ToUpper(strings.TrimPrefix(signPayload(payload, secret), "sha256="))

	if !VerifySignature(payload, signature, secret) {
		t.Fatal("expected uppercase hex signature to pass verification")
	}
}

func signPayload(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)

	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
