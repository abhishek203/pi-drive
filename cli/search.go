package cli

import (
	"fmt"
	"net/url"
	"os"
)

func runSearch(args []string) {
	parsed, err := parseCommandArgs(args, map[string]flagType{"type": stringFlag, "modified": stringFlag, "my-only": boolFlag, "shared-only": boolFlag})
	if err != nil {
		fatalf("%v", err)
	}
	if len(parsed.args) != 1 {
		fmt.Println("Usage: pidrive search <query> [--type ext] [--modified dur] [--my-only] [--shared-only]")
		os.Exit(1)
	}

	client, err := NewClient()
	if err != nil {
		fatalf("%v", err)
	}

	query := parsed.args[0]
	params := url.Values{}
	params.Set("q", query)
	if v := parsed.String("type", ""); v != "" {
		params.Set("type", v)
	}
	if v := parsed.String("modified", ""); v != "" {
		params.Set("modified", v)
	}
	if parsed.Bool("my-only") {
		params.Set("my_only", "true")
	}
	if parsed.Bool("shared-only") {
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
