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

func TestParseCredentials(t *testing.T) {
	creds, err := parseCredentials([]byte("api_key = \"pk_test\"\nserver = \"https://example.com\"\nmount_path = \"/tmp/drive\"\n"))
	if err != nil {
		t.Fatalf("parseCredentials() error = %v", err)
	}

	if creds.APIKey != "pk_test" || creds.Server != "https://example.com" || creds.Mount != "/tmp/drive" {
		t.Fatalf("parseCredentials() = %#v", creds)
	}
}

func TestEncodeCredentials(t *testing.T) {
	creds := &Credentials{
		APIKey: "pk_test",
		Server: "https://example.com",
		Mount:  "/tmp/drive",
	}

	got := string(encodeCredentials(creds))
	want := "api_key = \"pk_test\"\nserver = \"https://example.com\"\nmount_path = \"/tmp/drive\"\n"
	if got != want {
		t.Fatalf("encodeCredentials() = %q, want %q", got, want)
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
		{
			name:     "legacy darwin /drive uses default mount",
			mount:    "/drive",
			expected: "/drive", // adjusted below for darwin
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				creds: &Credentials{Mount: tt.mount},
			}

			got := c.MountPath()

			if tt.mount != "" {
				if runtime.GOOS == "darwin" && tt.mount == "/drive" {
					home, _ := os.UserHomeDir()
					expected := filepath.Join(home, "drive")
					if got != expected {
						t.Errorf("MountPath() = %q, want %q (darwin legacy)", got, expected)
					}
				} else if got != tt.expected {
					t.Errorf("MountPath() = %q, want %q", got, tt.expected)
				}
			} else {
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
