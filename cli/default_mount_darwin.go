//go:build darwin

package cli

import (
	"os"
	"path/filepath"
)

func defaultMountPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "drive")
}
