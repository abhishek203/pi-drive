package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func runRegister(args []string) {
	parsed, err := parseCommandArgs(args, map[string]flagType{"email": stringFlag, "name": stringFlag, "server": stringFlag})
	if err != nil {
		fatalf("%v", err)
	}

	email := parsed.String("email", "")
	name := parsed.String("name", "")
	server := parsed.String("server", "")
	if email == "" || name == "" || server == "" || len(parsed.args) != 0 {
		fmt.Println("Usage: pidrive register --email <email> --name <name> --server <url>")
		os.Exit(1)
	}

	client := NewClientWithServer(server)
	result, err := client.Post("/api/register", map[string]string{"email": email, "name": name})
	if err != nil {
		fatalf("Registration failed: %v", err)
	}

	apiKey, _ := result["api_key"].(string)
	if err := SaveCredentials(&Credentials{APIKey: apiKey, Server: server, Mount: "/drive"}); err != nil {
		fatalf("Failed to save credentials: %v", err)
	}

	fmt.Println("✓ Registered successfully!")
	fmt.Printf("  API Key: %s\n", apiKey)
	fmt.Printf("  Saved to: %s\n", credentialsPath())
	fmt.Println()
	fmt.Println("Check your email for the verification code, then run:")
	fmt.Printf("  pidrive verify --email %s --code <code>\n", email)
}

func runLogin(args []string) {
	parsed, err := parseCommandArgs(args, map[string]flagType{"email": stringFlag, "server": stringFlag})
	if err != nil {
		fatalf("%v", err)
	}

	email := parsed.String("email", "")
	server := parsed.String("server", "")
	if email == "" || len(parsed.args) != 0 {
		fmt.Println("Usage: pidrive login --email <email> [--server <url>]")
		os.Exit(1)
	}

	if server == "" {
		creds, err := LoadCredentials()
		if err == nil {
			server = creds.Server
		}
	}
	if server == "" {
		fmt.Println("--server is required (first time login)")
		os.Exit(1)
	}

	client := NewClientWithServer(server)
	if _, err := client.Post("/api/login", map[string]string{"email": email}); err != nil {
		fatalf("Login failed: %v", err)
	}

	fmt.Println("✓ Verification code sent to", email)
	fmt.Println()
	fmt.Print("Enter verification code: ")
	reader := bufio.NewReader(os.Stdin)
	code, _ := reader.ReadString('\n')
	code = strings.TrimSpace(code)

	result, err := client.Post("/api/verify", map[string]string{"email": email, "code": code})
	if err != nil {
		fatalf("Verification failed: %v", err)
	}

	apiKey, _ := result["api_key"].(string)
	if err := SaveCredentials(&Credentials{APIKey: apiKey, Server: server, Mount: "/drive"}); err != nil {
		fatalf("Failed to save credentials: %v", err)
	}

	fmt.Println("✓ Logged in successfully!")
}

func runVerify(args []string) {
	parsed, err := parseCommandArgs(args, map[string]flagType{"email": stringFlag, "code": stringFlag})
	if err != nil {
		fatalf("%v", err)
	}

	email := parsed.String("email", "")
	code := parsed.String("code", "")
	if email == "" || code == "" || len(parsed.args) != 0 {
		fmt.Println("Usage: pidrive verify --email <email> --code <code>")
		os.Exit(1)
	}

	creds, err := LoadCredentials()
	if err != nil {
		fatalf("%v", err)
	}

	client := NewClientWithServer(creds.Server)
	result, err := client.Post("/api/verify", map[string]string{"email": email, "code": code})
	if err != nil {
		fatalf("Verification failed: %v", err)
	}

	apiKey, _ := result["api_key"].(string)
	creds.APIKey = apiKey
	if err := SaveCredentials(creds); err != nil {
		fatalf("Failed to save credentials: %v", err)
	}

	fmt.Println("✓ Account verified!")
	fmt.Println()
	fmt.Println("Next: pidrive mount")
}

func runWhoami(args []string) {
	parsed, err := parseCommandArgs(args, nil)
	if err != nil {
		fatalf("%v", err)
	}
	if len(parsed.args) != 0 {
		fmt.Println("Usage: pidrive whoami")
		os.Exit(1)
	}

	client, err := NewClient()
	if err != nil {
		fatalf("%v", err)
	}

	result, err := client.Get("/api/me")
	if err != nil {
		fatalf("%v", err)
	}

	email, _ := result["email"].(string)
	name, _ := result["name"].(string)
	plan, _ := result["plan"].(string)
	usedBytes, _ := result["used_bytes"].(float64)
	quotaBytes, _ := result["quota_bytes"].(float64)

	fmt.Printf("%s (%s)\n", email, name)
	fmt.Printf("Plan: %s (%s / %s)\n", plan, formatBytes(int64(usedBytes)), formatBytes(int64(quotaBytes)))
	fmt.Printf("Server: %s\n", client.Server())
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
