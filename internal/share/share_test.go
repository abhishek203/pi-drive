package share

import (
	"testing"
)

func TestGenerateShareID(t *testing.T) {
	id1, err := generateShareID()
	if err != nil {
		t.Fatalf("generateShareID() failed: %v", err)
	}

	// Check length
	if len(id1) != 8 {
		t.Errorf("Share ID should be 8 chars, got: %d", len(id1))
	}

	// Check charset (lowercase alphanumeric)
	for _, c := range id1 {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			t.Errorf("Share ID should only contain lowercase alphanumeric, got: %c", c)
		}
	}

	// IDs should be unique
	id2, _ := generateShareID()
	if id1 == id2 {
		t.Error("Generated share IDs should be unique")
	}
}
