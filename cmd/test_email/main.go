package main

import (
	"fmt"
	"os"

	"github.com/pidrive/pidrive/internal/templates"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: test_email [verification|share|invite|admin]")
		os.Exit(1)
	}

	var html string
	var err error

	switch os.Args[1] {
	case "verification":
		html, err = templates.RenderVerification("847291")
	case "share":
		html, err = templates.RenderShareNotification("alice@company.com", "report.pdf")
	case "invite":
		html, err = templates.RenderShareInvite("alice@company.com", "report.pdf", "bob@company.com")
	case "admin":
		html, err = templates.RenderAdminNotification("newuser@example.com", "New User")
	default:
		fmt.Println("Unknown template")
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(html)
}
