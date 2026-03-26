package cli

import (
	"fmt"
	"net/url"
	"os"
)

func runActivity(args []string) {
	parsed, err := parseCommandArgs(args, map[string]flagType{"since": stringFlag, "type": stringFlag, "limit": intFlag})
	if err != nil {
		fatalf("%v", err)
	}
	if len(parsed.args) != 0 {
		fmt.Println("Usage: pidrive activity [--since dur] [--type action] [--limit n]")
		os.Exit(1)
	}

	client, err := NewClient()
	if err != nil {
		fatalf("%v", err)
	}

	limit, err := parsed.Int("limit", 50)
	if err != nil {
		fatalf("invalid --limit: %v", err)
	}

	params := url.Values{}
	if v := parsed.String("since", ""); v != "" {
		params.Set("since", v)
	}
	if v := parsed.String("type", ""); v != "" {
		params.Set("type", v)
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}

	path := "/api/activity"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	result, err := client.Get(path)
	if err != nil {
		fatalf("%v", err)
	}

	events, _ := result["events"].([]interface{})
	if len(events) == 0 {
		fmt.Println("No activity yet.")
		return
	}

	for _, e := range events {
		event := e.(map[string]interface{})
		createdAt, _ := event["created_at"].(string)
		action, _ := event["action"].(string)
		eventPath, _ := event["path"].(string)
		if len(createdAt) > 16 {
			createdAt = createdAt[11:16]
		}
		fmt.Printf("  %s  %-10s  %s\n", createdAt, action, eventPath)
	}
}
