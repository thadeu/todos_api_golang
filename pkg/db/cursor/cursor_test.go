package cursor

import (
	"os"
	"testing"
)

func TestEncodeDecodeCursor(t *testing.T) {
	// Set test secret key
	os.Setenv("CURSOR_SECRET_KEY", "test-secret-key-123")

	testDate := "2025-09-12T10:37:52.26483-03:00"
	testID := 123

	// Encode
	encoded := EncodeCursor(testDate, testID)
	t.Logf("Encoded cursor: %s", encoded)

	// Decode
	decodedDate, decodedID, err := DecodeCursor(encoded)

	if err != nil {
		t.Fatalf("Failed to decode cursor: %v", err)
	}

	// Verify
	if decodedDate != testDate {
		t.Errorf("Expected date %s, got %s", testDate, decodedDate)
	}

	if decodedID != testID {
		t.Errorf("Expected ID %d, got %d", testID, decodedID)
	}

	t.Logf("Successfully encoded and decoded: %s, %d -> %s, %d", testDate, testID, decodedDate, decodedID)
}

func TestDecodeInvalidCursor(t *testing.T) {
	os.Setenv("CURSOR_SECRET_KEY", "test-secret-key-123")

	// Test invalid format
	_, _, err := DecodeCursor("invalid-cursor")

	if err == nil {
		t.Error("Expected error for invalid cursor format")
	}

	// Test invalid signature
	invalidCursor := "eyJkYXRldGltZSI6IjIwMjUtMDktMTJUMTA6Mzc6NTItMDM6MDAifQ==.invalid-signature"
	_, _, err = DecodeCursor(invalidCursor)

	if err == nil {
		t.Error("Expected error for invalid signature")
	}
}
