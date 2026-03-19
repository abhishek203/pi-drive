package auth

import (
	"strings"
	"testing"
)

func TestGenerateAPIKey(t *testing.T) {
	key, hash, prefix, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey() failed: %v", err)
	}

	// Check key format
	if !strings.HasPrefix(key, "pk_") {
		t.Errorf("API key should start with 'pk_', got: %s", key[:min(10, len(key))])
	}
	if len(key) != 35 { // pk_ + 32 chars
		t.Errorf("API key should be 35 chars, got: %d", len(key))
	}

	// Check prefix
	if len(prefix) != 10 {
		t.Errorf("API key prefix should be 10 chars, got: %d", len(prefix))
	}
	if prefix != key[:10] {
		t.Errorf("Prefix should match first 10 chars of key")
	}

	// Check hash
	if len(hash) != 64 { // SHA-256 hex
		t.Errorf("Hash should be 64 hex chars, got: %d", len(hash))
	}
}

func TestGenerateVerificationCode(t *testing.T) {
	code, err := GenerateVerificationCode()
	if err != nil {
		t.Fatalf("GenerateVerificationCode() failed: %v", err)
	}

	if len(code) != 6 {
		t.Errorf("Verification code should be 6 digits, got: %d", len(code))
	}

	// Check all digits
	for _, c := range code {
		if c < '0' || c > '9' {
			t.Errorf("Verification code should only contain digits, got: %c", c)
		}
	}
}

func TestHashAPIKey(t *testing.T) {
	key := "pk_testkey12345678901234567890123"
	hash := HashAPIKey(key)

	// Hash should be consistent
	hash2 := HashAPIKey(key)
	if hash != hash2 {
		t.Error("HashAPIKey should return consistent results")
	}

	// Hash should be 64 hex chars (SHA-256)
	if len(hash) != 64 {
		t.Errorf("Hash should be 64 hex chars, got: %d", len(hash))
	}

	// Different keys should produce different hashes
	hash3 := HashAPIKey("pk_differentkey234567890123456789")
	if hash == hash3 {
		t.Error("Different keys should produce different hashes")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
