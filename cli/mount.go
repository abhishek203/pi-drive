package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func runMount(args []string) {
	fs := newFlagSet("mount")
	parseFlags(fs, args)
	if len(fs.Args()) != 0 {
		fmt.Println("Usage: pidrive mount")
		os.Exit(1)
	}

	client, err := NewClient()
	if err != nil {
		fatalf("%v", err)
	}

	fmt.Println("Connecting to server...")
	if _, err := client.Post("/api/mount", nil); err != nil {
		fatalf("%v", err)
	}

	drivePath := client.MountPath()
	os.MkdirAll(drivePath, 0755)
	if isMounted(drivePath) {
		fmt.Printf("✓ Already mounted at %s\n", drivePath)
		os.Exit(0)
	}

	fmt.Printf("Mounting at %s...\n", drivePath)
	if runtime.GOOS == "darwin" {
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
			fatalf("Mount failed: %v", err)
		}
	} else if runtime.GOOS == "linux" {
		if _, err := exec.LookPath("mount.davfs"); err != nil {
			fmt.Fprintln(os.Stderr, "✗ davfs2 is not installed")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "  Install:")
			fmt.Fprintln(os.Stderr, "    sudo apt install davfs2")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "  Then run 'pidrive mount' again.")
			os.Exit(1)
		}

		home, _ := os.UserHomeDir()
		davfsDir := home + "/.davfs2"
		os.MkdirAll(davfsDir, 0700)
		secretsFile := davfsDir + "/secrets"
		serverWebDAV := client.Server() + "/webdav"

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
			fatalf("Mount failed: %v", err)
		}
	} else {
		fatalf("Unsupported OS")
	}

	fmt.Println()
	fmt.Println("✓ Drive mounted!")
	fmt.Printf("  Your files:    %s/my/\n", drivePath)
	fmt.Printf("  Shared with you: %s/shared/\n", drivePath)
	fmt.Println()
	fmt.Println("Try:")
	fmt.Printf("  ls %s/my/\n", drivePath)
	fmt.Printf("  echo 'hello' > %s/my/test.txt\n", drivePath)
}

func runUnmount(args []string) {
	fs := newFlagSet("unmount")
	parseFlags(fs, args)
	if len(fs.Args()) != 0 {
		fmt.Println("Usage: pidrive unmount")
		os.Exit(1)
	}

	client, err := NewClient()
	if err != nil {
		fatalf("%v", err)
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
		fatalf("Unmount failed: %v", unmountErr)
	}

	client.Post("/api/unmount", nil)
	fmt.Println("✓ Drive unmounted")
}

func runStatus(args []string) {
	fs := newFlagSet("status")
	parseFlags(fs, args)
	if len(fs.Args()) != 0 {
		fmt.Println("Usage: pidrive status")
		os.Exit(1)
	}

	client, err := NewClient()
	if err != nil {
		fatalf("%v", err)
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
		return
	}

	email, _ := result["email"].(string)
	plan, _ := result["plan"].(string)
	usedBytes, _ := result["used_bytes"].(float64)
	quotaBytes, _ := result["quota_bytes"].(float64)
	fmt.Printf("Server:  ✓ %s\n", client.Server())
	fmt.Printf("Agent:   %s\n", email)
	fmt.Printf("Plan:    %s\n", plan)
	fmt.Printf("Storage: %s / %s\n", formatBytes(int64(usedBytes)), formatBytes(int64(quotaBytes)))
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
