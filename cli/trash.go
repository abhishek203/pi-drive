package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var trashCmd = &cobra.Command{
	Use:   "trash",
	Short: "List or manage deleted files",
	Run: func(cmd *cobra.Command, args []string) {
		empty, _ := cmd.Flags().GetBool("empty")
		if empty {
			emptyTrash()
			return
		}
		listTrash()
	},
}

var restoreCmd = &cobra.Command{
	Use:   "restore <path>",
	Short: "Restore a file from trash",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		_, err = client.Post("/api/trash/restore", map[string]string{
			"path": args[0],
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Restored %s\n", args[0])
	},
}

func listTrash() {
	client, err := NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %v\n", err)
		os.Exit(1)
	}

	result, err := client.Get("/api/trash")
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %v\n", err)
		os.Exit(1)
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
		fmt.Fprintf(os.Stderr, "✗ %v\n", err)
		os.Exit(1)
	}

	_, err = client.Delete("/api/trash")
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Trash emptied")
}

func init() {
	trashCmd.Flags().Bool("empty", false, "Permanently delete all trash")
}
