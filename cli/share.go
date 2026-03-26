package cli

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func runShare(args []string) {
	parsed, err := parseCommandArgs(args, map[string]flagType{"to": stringFlag, "link": boolFlag, "permission": stringFlag, "expires": stringFlag})
	if err != nil {
		fatalf("%v", err)
	}
	if len(parsed.args) != 1 {
		fmt.Println("Usage: pidrive share <path> --to <email>")
		fmt.Println("   or: pidrive share <path> --link")
		os.Exit(1)
	}

	client, err := NewClient()
	if err != nil {
		fatalf("%v", err)
	}

	filePath := parsed.args[0]
	drivePath := client.MountPath()
	myPath := filepath.Join(drivePath, "my")
	relPath, err := filepath.Rel(myPath, filePath)
	if err != nil || strings.HasPrefix(relPath, "..") {
		relPath = filePath
	}

	toEmail := parsed.String("to", "")
	asLink := parsed.Bool("link")
	permission := parsed.String("permission", "read")
	expires := parsed.String("expires", "")

	body := map[string]interface{}{"path": relPath}
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
		fatalf("%v", err)
	}

	if asLink {
		shareURL, _ := result["url"].(string)
		fmt.Printf("✓ %s\n", shareURL)
		return
	}
	fmt.Printf("✓ Shared %s with %s (%s)\n", relPath, toEmail, permission)
}

func runShared(args []string) {
	parsed, err := parseCommandArgs(args, nil)
	if err != nil {
		fatalf("%v", err)
	}
	if len(parsed.args) != 0 {
		fmt.Println("Usage: pidrive shared")
		os.Exit(1)
	}

	client, err := NewClient()
	if err != nil {
		fatalf("%v", err)
	}
	result, err := client.Get("/api/shared")
	if err != nil {
		fatalf("%v", err)
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
}

func runRevoke(args []string) {
	parsed, err := parseCommandArgs(args, nil)
	if err != nil {
		fatalf("%v", err)
	}
	if len(parsed.args) != 1 {
		fmt.Println("Usage: pidrive revoke <share-id or url>")
		os.Exit(1)
	}

	client, err := NewClient()
	if err != nil {
		fatalf("%v", err)
	}

	shareID := parsed.args[0]
	if strings.HasPrefix(shareID, "http") {
		u, err := url.Parse(shareID)
		if err == nil {
			parts := strings.Split(u.Path, "/")
			if len(parts) > 0 {
				shareID = parts[len(parts)-1]
			}
		}
	}

	if _, err := client.Delete("/api/share/" + shareID); err != nil {
		fatalf("%v", err)
	}
	fmt.Println("✓ Share revoked")
}

func runPull(args []string) {
	parsed, err := parseCommandArgs(args, nil)
	if err != nil {
		fatalf("%v", err)
	}
	if len(parsed.args) < 1 || len(parsed.args) > 2 {
		fmt.Println("Usage: pidrive pull <url> [destination]")
		os.Exit(1)
	}

	client, err := NewClient()
	if err != nil {
		fatalf("%v", err)
	}

	shareURL := parsed.args[0]
	dest := filepath.Join(client.MountPath(), "my", "incoming")
	if len(parsed.args) > 1 {
		dest = parsed.args[1]
	}

	os.MkdirAll(dest, 0755)
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
		fatalf("%v", err)
	}
	fmt.Printf("✓ Downloaded → %s\n", destFile)
}
