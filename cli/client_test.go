package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCredentialsPath(t *testing.T) {
	path := credentialsPath()

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home dir: %v", err)
	}

	expected := filepath.Join(home, ".pidrive", "credentials")
	if path != expected {
		t.Errorf("credentialsPath() = %q, want %q", path, expected)
	}
}

func TestClient_MountPath(t *testing.T) {
	tests := []struct {
		name     string
		mount    string
		expected string
	}{
		{
			name:     "custom mount path",
			mount:    "/custom/mount",
			expected: "/custom/mount",
		},
		{
			name:     "empty mount uses default",
			mount:    "",
			expected: "", // will be set based on OS
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				creds: &Credentials{Mount: tt.mount},
			}

			got := c.MountPath()

			if tt.mount != "" {
				if got != tt.expected {
					t.Errorf("MountPath() = %q, want %q", got, tt.expected)
				}
			} else {
				// Check OS-specific defaults
				if runtime.GOOS == "darwin" {
					home, _ := os.UserHomeDir()
					expected := filepath.Join(home, "drive")
					if got != expected {
						t.Errorf("MountPath() = %q, want %q (darwin)", got, expected)
					}
				} else {
					if got != "/drive" {
						t.Errorf("MountPath() = %q, want /drive (linux)", got)
					}
				}
			}
		})
	}
}

func TestClient_Server(t *testing.T) {
	c := &Client{
		creds: &Credentials{Server: "https://example.com"},
	}

	if got := c.Server(); got != "https://example.com" {
		t.Errorf("Server() = %q, want %q", got, "https://example.com")
	}
}

func TestNewClientWithServer(t *testing.T) {
	c := NewClientWithServer("https://test.com")

	if c.Server() != "https://test.com" {
		t.Errorf("Server() = %q, want %q", c.Server(), "https://test.com")
	}
}
