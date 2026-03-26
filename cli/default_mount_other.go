//go:build !darwin

package cli

func defaultMountPath() string {
	return "/drive"
}
