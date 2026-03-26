package cli

import (
	"fmt"
	"net/url"
	"os"
)

func runActivity(args []string) {
	fs := newFlagSet("activity")
	since := fs.String("since", "", "Show activity since (e.g. 1h, 7d)")
	actionType := fs.String("type", "", "Filter by action type (e.g. share, mount)")
	limit := fs.Int("limit", 50, "Maximum number of events")
	parseFlags(fs, args)
	if len(fs.Args()) != 0 {
		fmt.Println("Usage: pidrive activity [--since dur] [--type action] [--limit n]")
		os.Exit(1)
	}

	client, err := NewClient()
	if err != nil {
		fatalf("%v", err)
	}

	params := url.Values{}
	if *since != "" {
		params.Set("since", *since)
	}
	if *actionType != "" {
		params.Set("type", *actionType)
	}
	if *limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", *limit))
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
