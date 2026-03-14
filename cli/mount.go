package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var mountCmd = &cobra.Command{
	Use:   "mount",
	Short: "Mount your drive",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		// Call mount API to ensure agent dirs exist
		fmt.Println("Connecting to server...")
		_, err = client.Post("/api/mount", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		drivePath := client.MountPath()
		os.MkdirAll(drivePath, 0755)

		// Check if already mounted
		if isMounted(drivePath) {
			fmt.Printf("✓ Already mounted at %s\n", drivePath)
			os.Exit(0)
		}

		// WebDAV URL
		webdavURL := strings.Replace(client.Server(), "https://", "https://pidrive:"+client.creds.APIKey+"@", 1)
		webdavURL = strings.Replace(webdavURL, "http://", "http://pidrive:"+client.creds.APIKey+"@", 1)
		webdavURL += "/webdav"

		fmt.Printf("Mounting at %s...\n", drivePath)

		if runtime.GOOS == "darwin" {
			// macOS: use expect to automate mount_webdav -i
			if _, err := exec.LookPath("expect"); err != nil {
				fmt.Fprintln(os.Stderr, "✗ 'expect' is not installed")
				fmt.Fprintln(os.Stderr, "  Install: brew install expect")
				os.Exit(1)
			}

			serverWebDAV := client.Server() + "/webdav/"
			expectScript := fmt.Sprintf(`spawn mount_webdav -i %s %s
expect "Username:"
send "pidrive\r"
expect "Password:"
send "%s\r"
expect eof
`, serverWebDAV, drivePath, client.creds.APIKey)

			mountCmd := exec.Command("expect", "-c", expectScript)
			mountCmd.Stderr = os.Stderr
			mountCmd.Stdout = os.Stdout

			if err := mountCmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "✗ Mount failed: %v\n", err)
				os.Exit(1)
			}
		} else if runtime.GOOS == "linux" {
			// Linux: check for davfs2
			if _, err := exec.LookPath("mount.davfs"); err != nil {
				fmt.Fprintln(os.Stderr, "✗ davfs2 is not installed")
				fmt.Fprintln(os.Stderr, "")
				fmt.Fprintln(os.Stderr, "  Install:")
				fmt.Fprintln(os.Stderr, "    sudo apt install davfs2")
				fmt.Fprintln(os.Stderr, "")
				fmt.Fprintln(os.Stderr, "  Then run 'pidrive mount' again.")
				os.Exit(1)
			}

			// Write credentials to davfs2 secrets file
			home, _ := os.UserHomeDir()
			davfsDir := home + "/.davfs2"
			os.MkdirAll(davfsDir, 0700)
			secretsFile := davfsDir + "/secrets"
			serverWebDAV := client.Server() + "/webdav"

			// Read existing secrets, replace or append
			existing, _ := os.ReadFile(secretsFile)
			lines := strings.Split(string(existing), "\n")
			var newLines []string
			found := false
			for _, line := range lines {
				if strings.HasPrefix(line, serverWebDAV) {
					newLines = append(newLines, fmt.Sprintf("%s pidrive %s", serverWebDAV, client.creds.APIKey))
					found = true
				} else {
					newLines = append(newLines, line)
				}
			}
			if !found {
				newLines = append(newLines, fmt.Sprintf("%s pidrive %s", serverWebDAV, client.creds.APIKey))
			}
			os.WriteFile(secretsFile, []byte(strings.Join(newLines, "\n")), 0600)

			mountExec := exec.Command("sudo", "mount", "-t", "davfs", serverWebDAV, drivePath)
			mountExec.Stderr = os.Stderr
			mountExec.Stdout = os.Stdout
			mountExec.Stdin = os.Stdin

			if err := mountExec.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "✗ Mount failed: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Fprintln(os.Stderr, "✗ Unsupported OS")
			os.Exit(1)
		}

		fmt.Println()
		fmt.Println("✓ Drive mounted!")
		fmt.Printf("  Your files: %s/\n", drivePath)
		fmt.Println()
		fmt.Println("Try:")
		fmt.Printf("  ls %s/\n", drivePath)
		fmt.Printf("  echo 'hello' > %s/test.txt\n", drivePath)
	},
}

var unmountCmd = &cobra.Command{
	Use:   "unmount",
	Short: "Unmount your drive",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		drivePath := client.MountPath()
		fmt.Printf("Unmounting %s...\n", drivePath)

		var unmountErr error
		if runtime.GOOS == "linux" {
			unmountErr = exec.Command("sudo", "umount", drivePath).Run()
		} else {
			unmountErr = exec.Command("umount", drivePath).Run()
		}

		if unmountErr != nil {
			fmt.Fprintf(os.Stderr, "✗ Unmount failed: %v\n", unmountErr)
			os.Exit(1)
		}

		client.Post("/api/unmount", nil)
		fmt.Println("✓ Drive unmounted")
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show mount and connection status",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		drivePath := client.MountPath()

		if isMounted(drivePath) {
			fmt.Printf("Mount:   ✓ %s\n", drivePath)
		} else {
			fmt.Printf("Mount:   ✗ not mounted at %s\n", drivePath)
		}

		result, err := client.Get("/api/whoami")
		if err != nil {
			fmt.Printf("Server:  ✗ %v\n", err)
		} else {
			email, _ := result["email"].(string)
			plan, _ := result["plan"].(string)
			usedBytes, _ := result["used_bytes"].(float64)
			quotaBytes, _ := result["quota_bytes"].(float64)
			fmt.Printf("Server:  ✓ %s\n", client.Server())
			fmt.Printf("Agent:   %s\n", email)
			fmt.Printf("Plan:    %s\n", plan)
			fmt.Printf("Storage: %s / %s\n", formatBytes(int64(usedBytes)), formatBytes(int64(quotaBytes)))
		}
	},
}

func isMounted(path string) bool {
	if runtime.GOOS == "linux" {
		return exec.Command("mountpoint", "-q", path).Run() == nil
	}
	out, err := exec.Command("mount").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), path)
}
