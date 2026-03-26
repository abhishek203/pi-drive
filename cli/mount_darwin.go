//go:build darwin

package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func mountDrive(client *Client, drivePath string) error {
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
	return mountCmd.Run()
}

func unmountDrive(drivePath string) error {
	return exec.Command("umount", drivePath).Run()
}

func isMounted(path string) bool {
	out, err := exec.Command("mount").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), path)
}
