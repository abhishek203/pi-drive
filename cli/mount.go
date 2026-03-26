package cli

import (
	"fmt"
	"os"
)

func runMount(args []string) {
	parsed, err := parseCommandArgs(args, nil)
	if err != nil {
		fatalf("%v", err)
	}
	if len(parsed.args) != 0 {
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
	if err := mountDrive(client, drivePath); err != nil {
		fatalf("Mount failed: %v", err)
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
	parsed, err := parseCommandArgs(args, nil)
	if err != nil {
		fatalf("%v", err)
	}
	if len(parsed.args) != 0 {
		fmt.Println("Usage: pidrive unmount")
		os.Exit(1)
	}

	client, err := NewClient()
	if err != nil {
		fatalf("%v", err)
	}

	drivePath := client.MountPath()
	fmt.Printf("Unmounting %s...\n", drivePath)
	if err := unmountDrive(drivePath); err != nil {
		fatalf("Unmount failed: %v", err)
	}

	client.Post("/api/unmount", nil)
	fmt.Println("✓ Drive unmounted")
}

func runStatus(args []string) {
	parsed, err := parseCommandArgs(args, nil)
	if err != nil {
		fatalf("%v", err)
	}
	if len(parsed.args) != 0 {
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
