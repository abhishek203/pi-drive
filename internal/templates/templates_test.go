package templates

import (
	"strings"
	"testing"
)

func TestRenderVerification(t *testing.T) {
	html, err := RenderVerification("123456")
	if err != nil {
		t.Fatalf("RenderVerification() error: %v", err)
	}

	// Check that the code appears in the output
	if !strings.Contains(html, "123456") {
		t.Error("Rendered HTML should contain the verification code")
	}

	// Check that it's valid HTML
	if !strings.Contains(html, "<html") {
		t.Error("Rendered output should contain HTML")
	}
}

func TestRenderShareNotification(t *testing.T) {
	html, err := RenderShareNotification("alice@example.com", "report.pdf")
	if err != nil {
		t.Fatalf("RenderShareNotification() error: %v", err)
	}

	// Check that the from email appears
	if !strings.Contains(html, "alice@example.com") {
		t.Error("Rendered HTML should contain the from email")
	}

	// Check that the filename appears
	if !strings.Contains(html, "report.pdf") {
		t.Error("Rendered HTML should contain the filename")
	}
}

func TestRenderShareInvite(t *testing.T) {
	html, err := RenderShareInvite("alice@example.com", "data.csv", "bob@example.com")
	if err != nil {
		t.Fatalf("RenderShareInvite() error: %v", err)
	}

	// Check that both emails appear
	if !strings.Contains(html, "alice@example.com") {
		t.Error("Rendered HTML should contain the from email")
	}
	if !strings.Contains(html, "bob@example.com") {
		t.Error("Rendered HTML should contain the to email")
	}

	// Check that the filename appears
	if !strings.Contains(html, "data.csv") {
		t.Error("Rendered HTML should contain the filename")
	}
}

func TestRenderAdminNotification(t *testing.T) {
	html, err := RenderAdminNotification("newuser@example.com", "New User")
	if err != nil {
		t.Fatalf("RenderAdminNotification() error: %v", err)
	}

	// Check that the user info appears
	if !strings.Contains(html, "newuser@example.com") {
		t.Error("Rendered HTML should contain the user email")
	}
	if !strings.Contains(html, "New User") {
		t.Error("Rendered HTML should contain the user name")
	}
}
