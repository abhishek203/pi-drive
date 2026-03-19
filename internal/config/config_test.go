package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Save original env vars
	originalEnv := make(map[string]string)
	for _, key := range []string{
		"PIDRIVE_SERVER_URL",
		"PIDRIVE_PORT",
		"PIDRIVE_DATABASE_URL",
	} {
		originalEnv[key] = os.Getenv(key)
	}
	defer func() {
		for key, val := range originalEnv {
			if val == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, val)
			}
		}
	}()

	// Test default values
	os.Unsetenv("PIDRIVE_SERVER_URL")
	os.Unsetenv("PIDRIVE_PORT")
	cfg := Load()

	if cfg.ServerURL != "http://localhost:8080" {
		t.Errorf("Expected default ServerURL to be http://localhost:8080, got: %s", cfg.ServerURL)
	}
	if cfg.Port != "8080" {
		t.Errorf("Expected default Port to be 8080, got: %s", cfg.Port)
	}

	// Test custom values
	os.Setenv("PIDRIVE_SERVER_URL", "https://custom.example.com")
	os.Setenv("PIDRIVE_PORT", "3000")
	cfg = Load()

	if cfg.ServerURL != "https://custom.example.com" {
		t.Errorf("Expected ServerURL to be https://custom.example.com, got: %s", cfg.ServerURL)
	}
	if cfg.Port != "3000" {
		t.Errorf("Expected Port to be 3000, got: %s", cfg.Port)
	}
}

func TestEnvOr(t *testing.T) {
	// Test with env var set
	os.Setenv("TEST_ENV_VAR", "test_value")
	defer os.Unsetenv("TEST_ENV_VAR")

	result := envOr("TEST_ENV_VAR", "default")
	if result != "test_value" {
		t.Errorf("Expected 'test_value', got: %s", result)
	}

	// Test with env var not set
	result = envOr("NONEXISTENT_VAR", "default")
	if result != "default" {
		t.Errorf("Expected 'default', got: %s", result)
	}
}

func TestEnvIntOr(t *testing.T) {
	// Test with valid integer
	os.Setenv("TEST_INT_VAR", "42")
	defer os.Unsetenv("TEST_INT_VAR")

	result := envIntOr("TEST_INT_VAR", 10)
	if result != 42 {
		t.Errorf("Expected 42, got: %d", result)
	}

	// Test with invalid integer
	os.Setenv("TEST_INT_VAR", "not_a_number")
	result = envIntOr("TEST_INT_VAR", 10)
	if result != 10 {
		t.Errorf("Expected fallback 10, got: %d", result)
	}

	// Test with env var not set
	os.Unsetenv("TEST_INT_VAR")
	result = envIntOr("TEST_INT_VAR", 10)
	if result != 10 {
		t.Errorf("Expected fallback 10, got: %d", result)
	}
}
