package cli

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var shareCmd = &cobra.Command{
	Use:   "share <path>",
	Short: "Share a file with another agent or create a link",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		filePath := args[0]
		toEmail, _ := cmd.Flags().GetString("to")
		asLink, _ := cmd.Flags().GetBool("link")
		permission, _ := cmd.Flags().GetString("permission")
		expires, _ := cmd.Flags().GetString("expires")

		// Convert absolute path to relative path within drive
		drivePath := client.MountPath()
		myPath := filepath.Join(drivePath, "my")
		relPath, err := filepath.Rel(myPath, filePath)
		if err != nil || strings.HasPrefix(relPath, "..") {
			// Try as-is (user might have given relative path)
			relPath = filePath
		}

		body := map[string]interface{}{
			"path": relPath,
		}
		if toEmail != "" {
			body["to_email"] = toEmail
		}
		if asLink {
			body["link"] = true
		}
		if permission != "" {
			body["permission"] = permission
		}
		if expires != "" {
			body["expires"] = expires
		}

		if toEmail == "" && !asLink {
			fmt.Println("Usage: pidrive share <path> --to <email>")
			fmt.Println("   or: pidrive share <path> --link")
			os.Exit(1)
		}

		result, err := client.Post("/api/share", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		if asLink {
			shareURL, _ := result["url"].(string)
			fmt.Printf("✓ %s\n", shareURL)
		} else {
			fmt.Printf("✓ Shared %s with %s (%s)\n", relPath, toEmail, permission)
		}
	},
}

var sharedCmd = &cobra.Command{
	Use:   "shared",
	Short: "List all shares",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		result, err := client.Get("/api/shared")
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		byMe, _ := result["shared_by_me"].([]interface{})
		withMe, _ := result["shared_with_me"].([]interface{})

		if len(byMe) == 0 && len(withMe) == 0 {
			fmt.Println("No shares yet.")
			return
		}

		if len(byMe) > 0 {
			fmt.Println("SHARED BY ME:")
			for _, s := range byMe {
				share := s.(map[string]interface{})
				path, _ := share["source_path"].(string)
				shareType, _ := share["share_type"].(string)
				if shareType == "link" {
					shareURL, _ := share["url"].(string)
					fmt.Printf("  %-30s → link: %s\n", path, shareURL)
				} else {
					targetEmail, _ := share["target_email"].(string)
					perm, _ := share["permission"].(string)
					fmt.Printf("  %-30s → %s (%s)\n", path, targetEmail, perm)
				}
			}
			fmt.Println()
		}

		if len(withMe) > 0 {
			fmt.Println("SHARED WITH ME:")
			for _, s := range withMe {
				share := s.(map[string]interface{})
				path, _ := share["source_path"].(string)
				ownerEmail, _ := share["owner_email"].(string)
				perm, _ := share["permission"].(string)
				fmt.Printf("  %-30s ← %s (%s)\n", path, ownerEmail, perm)
			}
		}
	},
}

var revokeCmd = &cobra.Command{
	Use:   "revoke <share-id or url>",
	Short: "Revoke a share",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		shareID := args[0]

		// If it's a URL, extract the share ID
		if strings.HasPrefix(shareID, "http") {
			u, err := url.Parse(shareID)
			if err == nil {
				parts := strings.Split(u.Path, "/")
				if len(parts) > 0 {
					shareID = parts[len(parts)-1]
				}
			}
		}

		_, err = client.Delete("/api/share/" + shareID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ Share revoked")
	},
}

var pullCmd = &cobra.Command{
	Use:   "pull <url> [destination]",
	Short: "Download a shared file",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		shareURL := args[0]
		dest := filepath.Join(client.MountPath(), "my", "incoming")
		if len(args) > 1 {
			dest = args[1]
		}

		// Ensure dest directory exists
		os.MkdirAll(dest, 0755)

		// Extract filename from URL or use default
		filename := "downloaded-file"
		u, err := url.Parse(shareURL)
		if err == nil {
			parts := strings.Split(u.Path, "/")
			if len(parts) > 0 {
				filename = parts[len(parts)-1]
			}
		}

		destFile := filepath.Join(dest, filename)
		if info, err := os.Stat(dest); err == nil && !info.IsDir() {
			destFile = dest
		}

		fmt.Printf("Downloading to %s...\n", destFile)
		if err := client.DownloadFile(shareURL, destFile); err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Downloaded → %s\n", destFile)
	},
}

func init() {
	shareCmd.Flags().String("to", "", "Email of agent to share with")
	shareCmd.Flags().Bool("link", false, "Create a shareable link")
	shareCmd.Flags().String("permission", "read", "Permission: read or write")
	shareCmd.Flags().String("expires", "", "Expiry duration (e.g. 7d, 24h)")
}
