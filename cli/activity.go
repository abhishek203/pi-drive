package cli

import (
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

var activityCmd = &cobra.Command{
	Use:   "activity",
	Short: "Show recent activity",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		since, _ := cmd.Flags().GetString("since")
		actionType, _ := cmd.Flags().GetString("type")
		limit, _ := cmd.Flags().GetInt("limit")

		params := url.Values{}
		if since != "" {
			params.Set("since", since)
		}
		if actionType != "" {
			params.Set("type", actionType)
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
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
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

			// Truncate timestamp to time only
			if len(createdAt) > 16 {
				createdAt = createdAt[11:16]
			}

			fmt.Printf("  %s  %-10s  %s\n", createdAt, action, eventPath)
		}
	},
}

func init() {
	activityCmd.Flags().String("since", "", "Show activity since (e.g. 1h, 7d)")
	activityCmd.Flags().String("type", "", "Filter by action type (e.g. share, mount)")
	activityCmd.Flags().Int("limit", 50, "Maximum number of events")
}
