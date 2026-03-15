package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type EmailService struct {
	ResendAPIKey string
	FromEmail    string
}

func NewEmailService(resendAPIKey, fromEmail string) *EmailService {
	return &EmailService{
		ResendAPIKey: resendAPIKey,
		FromEmail:    fromEmail,
	}
}

func (e *EmailService) SendVerificationCode(toEmail, code string) error {
	// Dev mode: no API key configured, just log
	if e.ResendAPIKey == "" {
		log.Printf("[DEV] Verification code for %s: %s", toEmail, code)
		return nil
	}

	payload := map[string]string{
		"from":    fmt.Sprintf("pidrive <%s>", e.FromEmail),
		"to":      toEmail,
		"subject": "Your pidrive verification code",
		"text":    fmt.Sprintf("Your pidrive verification code is: %s\n\nThis code expires in 15 minutes.", code),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+e.ResendAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend API error (%d): %s", resp.StatusCode, string(respBody))
	}

	log.Printf("[EMAIL] Verification code sent to %s", toEmail)
	return nil
}

func (e *EmailService) SendShareNotification(toEmail, fromEmail, filename string) error {
	if e.ResendAPIKey == "" {
		log.Printf("[DEV] Share notification: %s shared %s with %s", fromEmail, filename, toEmail)
		return nil
	}

	payload := map[string]string{
		"from":    fmt.Sprintf("pidrive <%s>", e.FromEmail),
		"to":      toEmail,
		"subject": fmt.Sprintf("%s shared \"%s\" with you", fromEmail, filename),
		"text": fmt.Sprintf(`%s shared a file with you on pidrive.

File: %s

View it in your drive:
  pidrive mount
  cat /drive/shared/%s/%s

Or run:
  pidrive shared
`, fromEmail, filename, fromEmail, filename),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+e.ResendAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend API error (%d): %s", resp.StatusCode, string(respBody))
	}

	log.Printf("[EMAIL] Share notification sent to %s (from %s, file: %s)", toEmail, fromEmail, filename)
	return nil
}

func (e *EmailService) SendShareInvite(toEmail, fromEmail, filename string) error {
	if e.ResendAPIKey == "" {
		log.Printf("[DEV] Share invite: %s shared %s with %s (not registered)", fromEmail, filename, toEmail)
		return nil
	}

	payload := map[string]string{
		"from":    fmt.Sprintf("pidrive <%s>", e.FromEmail),
		"to":      toEmail,
		"subject": fmt.Sprintf("%s shared \"%s\" with you on pidrive", fromEmail, filename),
		"text": fmt.Sprintf(`%s shared a file with you on pidrive.

File: %s

To access it, sign up for pidrive (free):

  curl -sSL https://pidrive.ressl.ai/install.sh | bash
  pidrive register --email %s --name "My Agent" --server https://pidrive.ressl.ai

Once registered, the shared file will appear in your drive automatically.
`, fromEmail, filename, toEmail),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+e.ResendAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend API error (%d): %s", resp.StatusCode, string(respBody))
	}

	log.Printf("[EMAIL] Share invite sent to %s (from %s, file: %s)", toEmail, fromEmail, filename)
	return nil
}
