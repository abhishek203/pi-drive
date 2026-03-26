package cli

import (
	"fmt"
	"os"
)

func runTrash(args []string) {
	fs := newFlagSet("trash")
	empty := fs.Bool("empty", false, "Permanently delete all trash")
	parseFlags(fs, args)
	if len(fs.Args()) != 0 {
		fmt.Println("Usage: pidrive trash [--empty]")
		os.Exit(1)
	}
	if *empty {
		emptyTrash()
		return
	}
	listTrash()
}

func runRestore(args []string) {
	fs := newFlagSet("restore")
	parseFlags(fs, args)
	if len(fs.Args()) != 1 {
		fmt.Println("Usage: pidrive restore <path>")
		os.Exit(1)
	}

	client, err := NewClient()
	if err != nil {
		fatalf("%v", err)
	}

	if _, err := client.Post("/api/trash/restore", map[string]string{"path": fs.Args()[0]}); err != nil {
		fatalf("%v", err)
	}

	fmt.Printf("✓ Restored %s\n", fs.Args()[0])
}

func listTrash() {
	client, err := NewClient()
	if err != nil {
		fatalf("%v", err)
	}

	result, err := client.Get("/api/trash")
	if err != nil {
		fatalf("%v", err)
	}

	items, _ := result["items"].([]interface{})
	if len(items) == 0 {
		fmt.Println("Trash is empty.")
		return
	}

	fmt.Println("TRASH:")
	for _, item := range items {
		i := item.(map[string]interface{})
		path, _ := i["path"].(string)
		deletedAt, _ := i["deleted_at"].(string)
		recoverableUntil, _ := i["recoverable_until"].(string)
		if len(deletedAt) > 10 {
			deletedAt = deletedAt[:10]
		}
		if len(recoverableUntil) > 10 {
			recoverableUntil = recoverableUntil[:10]
		}
		fmt.Printf("  %-30s  deleted %s  recoverable until %s\n", path, deletedAt, recoverableUntil)
	}
	fmt.Println()
	fmt.Println("Restore: pidrive restore <path>")
	fmt.Println("Empty:   pidrive trash --empty")
}

func emptyTrash() {
	client, err := NewClient()
	if err != nil {
		fatalf("%v", err)
	}

	if _, err := client.Delete("/api/trash"); err != nil {
		fatalf("%v", err)
	}

	fmt.Println("✓ Trash emptied")
}
