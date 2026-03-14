package cli

import (
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search across your files and shared files",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		query := args[0]
		fileType, _ := cmd.Flags().GetString("type")
		modified, _ := cmd.Flags().GetString("modified")
		myOnly, _ := cmd.Flags().GetBool("my-only")
		sharedOnly, _ := cmd.Flags().GetBool("shared-only")

		params := url.Values{}
		params.Set("q", query)
		if fileType != "" {
			params.Set("type", fileType)
		}
		if modified != "" {
			params.Set("modified", modified)
		}
		if myOnly {
			params.Set("my_only", "true")
		}
		if sharedOnly {
			params.Set("shared_only", "true")
		}

		result, err := client.Get("/api/search?" + params.Encode())
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		results, _ := result["results"].([]interface{})
		count := int(result["count"].(float64))

		if count == 0 {
			fmt.Println("No results found.")
			return
		}

		fmt.Printf("%d result(s) for \"%s\":\n\n", count, query)
		for _, r := range results {
			res := r.(map[string]interface{})
			path, _ := res["path"].(string)
			snippet, _ := res["snippet"].(string)
			sizeBytes, _ := res["size_bytes"].(float64)
			fmt.Printf("  %s (%s)\n", path, formatBytes(int64(sizeBytes)))
			if snippet != "" {
				fmt.Printf("    %s\n", snippet)
			}
			fmt.Println()
		}
	},
}

func init() {
	searchCmd.Flags().String("type", "", "Filter by file extension (e.g. csv,txt)")
	searchCmd.Flags().String("modified", "", "Filter by modification time (e.g. 7d, 24h)")
	searchCmd.Flags().Bool("my-only", false, "Search only your own files")
	searchCmd.Flags().Bool("shared-only", false, "Search only shared files")
}
