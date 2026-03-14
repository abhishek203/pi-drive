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
