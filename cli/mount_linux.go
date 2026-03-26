//go:build linux

package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func mountDrive(client *Client, drivePath string) error {
	if _, err := exec.LookPath("mount.davfs"); err != nil {
		fmt.Fprintln(os.Stderr, "✗ davfs2 is not installed")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "  Install:")
		fmt.Fprintln(os.Stderr, "    sudo apt install davfs2")
		fmt.Fprintln(os.Stderr)
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
	return mountExec.Run()
}

func unmountDrive(drivePath string) error {
	return exec.Command("sudo", "umount", drivePath).Run()
}

func isMounted(path string) bool {
	return exec.Command("mountpoint", "-q", path).Run() == nil
}
