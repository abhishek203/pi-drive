package search

import (
	"testing"
)

func TestIndexableExtensions(t *testing.T) {
	// Test that common text file extensions are indexable
	textExts := []string{".txt", ".md", ".csv", ".json", ".yaml", ".yml"}
	for _, ext := range textExts {
		if !indexableExtensions[ext] {
			t.Errorf("Expected %s to be indexable", ext)
		}
	}

	// Test that code extensions are indexable
	codeExts := []string{".py", ".js", ".ts", ".go", ".java", ".rb"}
	for _, ext := range codeExts {
		if !indexableExtensions[ext] {
			t.Errorf("Expected %s to be indexable", ext)
		}
	}

	// Test that binary/image extensions are not indexable
	nonIndexable := []string{".jpg", ".png", ".gif", ".exe", ".zip"}
	for _, ext := range nonIndexable {
		if indexableExtensions[ext] {
			t.Errorf("Expected %s to NOT be indexable", ext)
		}
	}
}

func TestMaxFileSize(t *testing.T) {
	expected := int64(50 * 1024 * 1024) // 50 MB
	if maxFileSize != expected {
		t.Errorf("Expected maxFileSize to be %d, got %d", expected, maxFileSize)
	}
}
