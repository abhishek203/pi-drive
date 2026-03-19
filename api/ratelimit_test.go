package api

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := newRateLimiter(3, time.Minute)

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		if !rl.allow("test-key") {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 4th request should be denied
	if rl.allow("test-key") {
		t.Error("Request 4 should be denied")
	}

	// Different key should still be allowed
	if !rl.allow("different-key") {
		t.Error("Different key should be allowed")
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	// Use a very short window for testing
	rl := &rateLimiter{
		attempts: make(map[string][]time.Time),
		limit:    2,
		window:   10 * time.Millisecond,
	}

	// First 2 requests allowed
	rl.allow("key")
	rl.allow("key")

	// 3rd should be denied
	if rl.allow("key") {
		t.Error("Should be denied before window expires")
	}

	// Wait for window to expire
	time.Sleep(15 * time.Millisecond)

	// Should be allowed again
	if !rl.allow("key") {
		t.Error("Should be allowed after window expires")
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name:       "X-Forwarded-For",
			headers:    map[string]string{"X-Forwarded-For": "1.2.3.4"},
			remoteAddr: "127.0.0.1:12345",
			expected:   "1.2.3.4",
		},
		{
			name:       "X-Real-IP",
			headers:    map[string]string{"X-Real-IP": "5.6.7.8"},
			remoteAddr: "127.0.0.1:12345",
			expected:   "5.6.7.8",
		},
		{
			name:       "RemoteAddr fallback",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.1:54321",
			expected:   "192.168.1.1:54321",
		},
		{
			name:       "X-Forwarded-For priority",
			headers:    map[string]string{"X-Forwarded-For": "1.1.1.1", "X-Real-IP": "2.2.2.2"},
			remoteAddr: "3.3.3.3:12345",
			expected:   "1.1.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			req.RemoteAddr = tt.remoteAddr

			got := getClientIP(req)
			if got != tt.expected {
				t.Errorf("getClientIP() = %q, want %q", got, tt.expected)
			}
		})
	}
}
