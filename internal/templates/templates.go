package templates

import (
	"bytes"
	"embed"
	"html/template"
)

//go:embed *.html
var templateFS embed.FS

// parseTemplate parses base.html with a specific template file
func parseTemplate(name string) (*template.Template, error) {
	return template.ParseFS(templateFS, "base.html", name)
}

// VerificationData holds data for the verification email template
type VerificationData struct {
	Code string
}

// ShareNotificationData holds data for share notification email template
type ShareNotificationData struct {
	FromEmail string
	Filename  string
}

// ShareInviteData holds data for share invite email template
type ShareInviteData struct {
	FromEmail string
	Filename  string
	ToEmail   string
}

// AdminNotificationData holds data for admin notification email template
type AdminNotificationData struct {
	UserEmail string
	UserName  string
}

// RenderVerification renders the verification email HTML
func RenderVerification(code string) (string, error) {
	tmpl, err := parseTemplate("verification.html")
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "base", VerificationData{Code: code})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderShareNotification renders the share notification email HTML
func RenderShareNotification(fromEmail, filename string) (string, error) {
	tmpl, err := parseTemplate("share_notification.html")
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	data := ShareNotificationData{FromEmail: fromEmail, Filename: filename}
	err = tmpl.ExecuteTemplate(&buf, "base", data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderShareInvite renders the share invite email HTML
func RenderShareInvite(fromEmail, filename, toEmail string) (string, error) {
	tmpl, err := parseTemplate("share_invite.html")
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	data := ShareInviteData{FromEmail: fromEmail, Filename: filename, ToEmail: toEmail}
	err = tmpl.ExecuteTemplate(&buf, "base", data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderAdminNotification renders the admin notification email HTML
func RenderAdminNotification(userEmail, userName string) (string, error) {
	tmpl, err := parseTemplate("admin_notification.html")
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	data := AdminNotificationData{UserEmail: userEmail, UserName: userName}
	err = tmpl.ExecuteTemplate(&buf, "base", data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
