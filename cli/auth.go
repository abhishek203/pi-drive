package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new agent account",
	Run: func(cmd *cobra.Command, args []string) {
		email, _ := cmd.Flags().GetString("email")
		name, _ := cmd.Flags().GetString("name")
		server, _ := cmd.Flags().GetString("server")

		if email == "" || name == "" || server == "" {
			fmt.Println("Usage: pidrive register --email <email> --name <name> --server <url>")
			os.Exit(1)
		}

		client := NewClientWithServer(server)
		result, err := client.Post("/api/register", map[string]string{
			"email": email,
			"name":  name,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Registration failed: %v\n", err)
			os.Exit(1)
		}

		apiKey, _ := result["api_key"].(string)

		// Save credentials
		if err := SaveCredentials(&Credentials{
			APIKey: apiKey,
			Server: server,
			Mount:  "/drive",
		}); err != nil {
			fmt.Fprintf(os.Stderr, "✗ Failed to save credentials: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ Registered successfully!")
		fmt.Printf("  API Key: %s\n", apiKey)
		fmt.Printf("  Saved to: %s\n", credentialsPath())
		fmt.Println()
		fmt.Println("Check your email for the verification code, then run:")
		fmt.Printf("  pidrive verify --email %s --code <code>\n", email)
	},
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to an existing account",
	Run: func(cmd *cobra.Command, args []string) {
		email, _ := cmd.Flags().GetString("email")
		server, _ := cmd.Flags().GetString("server")

		if email == "" {
			fmt.Println("Usage: pidrive login --email <email> [--server <url>]")
			os.Exit(1)
		}

		// Try to load existing server from credentials
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
		_, err := client.Post("/api/login", map[string]string{
			"email": email,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Login failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ Verification code sent to", email)
		fmt.Println()

		// Prompt for code
		fmt.Print("Enter verification code: ")
		reader := bufio.NewReader(os.Stdin)
		code, _ := reader.ReadString('\n')
		code = strings.TrimSpace(code)

		result, err := client.Post("/api/verify", map[string]string{
			"email": email,
			"code":  code,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Verification failed: %v\n", err)
			os.Exit(1)
		}

		apiKey, _ := result["api_key"].(string)
		if err := SaveCredentials(&Credentials{
			APIKey: apiKey,
			Server: server,
			Mount:  "/drive",
		}); err != nil {
			fmt.Fprintf(os.Stderr, "✗ Failed to save credentials: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ Logged in successfully!")
	},
}

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify your account with the code from email",
	Run: func(cmd *cobra.Command, args []string) {
		email, _ := cmd.Flags().GetString("email")
		code, _ := cmd.Flags().GetString("code")

		if email == "" || code == "" {
			fmt.Println("Usage: pidrive verify --email <email> --code <code>")
			os.Exit(1)
		}

		// Load server from credentials
		creds, err := LoadCredentials()
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		client := NewClientWithServer(creds.Server)
		result, err := client.Post("/api/verify", map[string]string{
			"email": email,
			"code":  code,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Verification failed: %v\n", err)
			os.Exit(1)
		}

		apiKey, _ := result["api_key"].(string)
		creds.APIKey = apiKey
		SaveCredentials(creds)

		fmt.Println("✓ Account verified!")
		fmt.Println()
		fmt.Println("Next: pidrive mount")
	},
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current agent info",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		result, err := client.Get("/api/me")
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		email, _ := result["email"].(string)
		name, _ := result["name"].(string)
		plan, _ := result["plan"].(string)
		usedBytes, _ := result["used_bytes"].(float64)
		quotaBytes, _ := result["quota_bytes"].(float64)

		fmt.Printf("%s (%s)\n", email, name)
		fmt.Printf("Plan: %s (%s / %s)\n", plan, formatBytes(int64(usedBytes)), formatBytes(int64(quotaBytes)))
		fmt.Printf("Server: %s\n", client.Server())
	},
}

func init() {
	registerCmd.Flags().String("email", "", "Email address")
	registerCmd.Flags().String("name", "", "Agent name")
	registerCmd.Flags().String("server", "", "pidrive server URL")

	loginCmd.Flags().String("email", "", "Email address")
	loginCmd.Flags().String("server", "", "pidrive server URL")

	verifyCmd.Flags().String("email", "", "Email address")
	verifyCmd.Flags().String("code", "", "Verification code")
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
