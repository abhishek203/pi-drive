package api

import (
	"os"
	"testing"
)

func TestCleanPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "/"},
		{"/", "/"},
		{".", "/"},
		{"foo", "/foo"},
		{"/foo", "/foo"},
		{"/foo/bar", "/foo/bar"},
		{"/foo//bar", "/foo/bar"},
		{"/foo/../bar", "/bar"},
		{"foo/bar", "/foo/bar"},
		{"/foo/./bar", "/foo/bar"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := cleanPath(tt.input)
			if got != tt.expected {
				t.Errorf("cleanPath(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsSharedPath(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/shared", true},
		{"/shared/", true},
		{"/shared/email@example.com", true},
		{"/shared/email@example.com/file.txt", true},
		{"/my", false},
		{"/my/file.txt", false},
		{"/", false},
		{"", false},
		{"/share", false}, // not the same as /shared
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isSharedPath(tt.path)
			if got != tt.expected {
				t.Errorf("isSharedPath(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestVirtualDir(t *testing.T) {
	vd := &virtualDir{
		name: "test",
		entries: []os.FileInfo{
			&dirInfo{name: "file1"},
			&dirInfo{name: "file2"},
		},
	}

	// Test Stat
	info, err := vd.Stat()
	if err != nil {
		t.Fatalf("Stat() error: %v", err)
	}
	if !info.IsDir() {
		t.Error("Stat() should return a directory")
	}
	if info.Name() != "test" {
		t.Errorf("Stat().Name() = %q, want %q", info.Name(), "test")
	}

	// Test Close
	if err := vd.Close(); err != nil {
		t.Errorf("Close() error: %v", err)
	}

	// Test Read (should error for directory)
	_, err = vd.Read(make([]byte, 10))
	if err == nil {
		t.Error("Read() should return an error for directory")
	}

	// Test Write (should error for directory)
	_, err = vd.Write([]byte("test"))
	if err == nil {
		t.Error("Write() should return an error for directory")
	}

	// Test Seek (should error for directory)
	_, err = vd.Seek(0, 0)
	if err == nil {
		t.Error("Seek() should return an error for directory")
	}
}

func TestDirInfo(t *testing.T) {
	di := &dirInfo{name: "testdir"}

	if di.Name() != "testdir" {
		t.Errorf("Name() = %q, want %q", di.Name(), "testdir")
	}
	if di.Size() != 0 {
		t.Errorf("Size() = %d, want 0", di.Size())
	}
	if !di.IsDir() {
		t.Error("IsDir() should return true")
	}
	if di.Mode().IsDir() != true {
		t.Error("Mode().IsDir() should be true")
	}
	if di.Sys() != nil {
		t.Error("Sys() should return nil")
	}
}
