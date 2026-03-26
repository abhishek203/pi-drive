//go:build !darwin && !linux

package cli

import "fmt"

func mountDrive(client *Client, drivePath string) error {
	return fmt.Errorf("unsupported OS")
}

func unmountDrive(drivePath string) error {
	return fmt.Errorf("unsupported OS")
}

func isMounted(path string) bool {
	return false
}
