package cli

import (
	"fmt"
	"net/url"
	"os"
)

func runSearch(args []string) {
	fs := newFlagSet("search")
	fileType := fs.String("type", "", "Filter by file extension (e.g. csv,txt)")
	modified := fs.String("modified", "", "Filter by modification time (e.g. 7d, 24h)")
	myOnly := fs.Bool("my-only", false, "Search only your own files")
	sharedOnly := fs.Bool("shared-only", false, "Search only shared files")
	parseFlags(fs, args)
	if len(fs.Args()) != 1 {
		fmt.Println("Usage: pidrive search <query> [--type ext] [--modified dur] [--my-only] [--shared-only]")
		os.Exit(1)
	}

	client, err := NewClient()
	if err != nil {
		fatalf("%v", err)
	}

	query := fs.Args()[0]
	params := url.Values{}
	params.Set("q", query)
	if *fileType != "" {
		params.Set("type", *fileType)
	}
	if *modified != "" {
		params.Set("modified", *modified)
	}
	if *myOnly {
		params.Set("my_only", "true")
	}
	if *sharedOnly {
		params.Set("shared_only", "true")
	}

	result, err := client.Get("/api/search?" + params.Encode())
	if err != nil {
		fatalf("%v", err)
	}

	results, _ := result["results"].([]interface{})
	count := int(result["count"].(float64))
	if count == 0 {
		fmt.Println("No results found.")
		return
	}

	fmt.Printf("%d result(s) for %q:\n\n", count, query)
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
}
