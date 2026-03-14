package cli

import (
	"fmt"
	"net/url"
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

		// Check sshfs is installed
		if _, err := exec.LookPath("sshfs"); err != nil {
			fmt.Fprintln(os.Stderr, "✗ sshfs is not installed")
			fmt.Fprintln(os.Stderr, "")
			if runtime.GOOS == "linux" {
				fmt.Fprintln(os.Stderr, "  Install:")
				fmt.Fprintln(os.Stderr, "    sudo apt install sshfs")
			} else if runtime.GOOS == "darwin" {
				fmt.Fprintln(os.Stderr, "  macOS requires macFUSE + sshfs (one-time setup):")
				fmt.Fprintln(os.Stderr, "")
				fmt.Fprintln(os.Stderr, "  Step 1: Install macFUSE")
				fmt.Fprintln(os.Stderr, "    brew install --cask macfuse")
				fmt.Fprintln(os.Stderr, "")
				fmt.Fprintln(os.Stderr, "  Step 2: Allow the kernel extension")
				fmt.Fprintln(os.Stderr, "    System Settings → Privacy & Security → scroll down → Allow")
				fmt.Fprintln(os.Stderr, "")
				fmt.Fprintln(os.Stderr, "  Step 3: Reboot your Mac")
				fmt.Fprintln(os.Stderr, "")
				fmt.Fprintln(os.Stderr, "  Step 4: Install sshfs")
				fmt.Fprintln(os.Stderr, "    brew install gromgit/fuse/sshfs-mac")
				fmt.Fprintln(os.Stderr, "")
				fmt.Fprintln(os.Stderr, "  Then run 'pidrive mount' again.")
			} else {
				fmt.Fprintln(os.Stderr, "  Install sshfs for your platform and try again.")
			}
			os.Exit(1)
		}

		// Call mount API to ensure agent dirs exist
		fmt.Println("Connecting to server...")
		result, err := client.Post("/api/mount", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		agentID, _ := result["agent_id"].(string)

		drivePath := client.MountPath()
		os.MkdirAll(drivePath, 0755)

		// Check if already mounted
		if isMounted(drivePath) {
			fmt.Printf("✓ Already mounted at %s\n", drivePath)
			os.Exit(0)
		}

		// Parse server URL to get host
		serverURL := client.Server()
		u, err := url.Parse(serverURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Invalid server URL: %v\n", err)
			os.Exit(1)
		}
		host := u.Hostname()

		// SFTP port
		sftpPort := "2022"

		fmt.Printf("Mounting at %s...\n", drivePath)

		// Mount via sshfs
		// Username is agent ID, password is API key
		sshfsArgs := []string{
			fmt.Sprintf("%s@%s:/", agentID, host),
			drivePath,
			"-p", sftpPort,
			"-o", "password_stdin",
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "reconnect",
			"-o", "ServerAliveInterval=15",
			"-o", "ServerAliveCountMax=3",
		}

		if runtime.GOOS == "darwin" {
			sshfsArgs = append(sshfsArgs, "-o", "volname=pidrive")
		}

		sshfsCmd := exec.Command("sshfs", sshfsArgs...)
		sshfsCmd.Stdin = strings.NewReader(client.creds.APIKey)
		sshfsCmd.Stderr = os.Stderr

		if err := sshfsCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "✗ Mount failed: %v\n", err)
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
			unmountErr = exec.Command("fusermount", "-u", drivePath).Run()
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

func init() {
	mountCmd.Flags().String("path", "/drive", "Mount path")
}
