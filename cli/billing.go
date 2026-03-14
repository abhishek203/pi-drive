package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var usageCmd = &cobra.Command{
	Use:   "usage",
	Short: "Show storage and bandwidth usage",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		result, err := client.Get("/api/usage")
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		usedBytes := int64(result["used_bytes"].(float64))
		quotaBytes := int64(result["quota_bytes"].(float64))
		plan, _ := result["plan"].(string)
		bwToday := int64(result["bandwidth_today"].(float64))
		bwMonth := int64(result["bandwidth_this_month"].(float64))

		pct := float64(0)
		if quotaBytes > 0 {
			pct = float64(usedBytes) / float64(quotaBytes) * 100
		}

		fmt.Printf("Storage:    %s / %s (%.0f%%)\n", formatBytes(usedBytes), formatBytes(quotaBytes), pct)
		fmt.Printf("Bandwidth:  %s today, %s this month\n", formatBytes(bwToday), formatBytes(bwMonth))
		fmt.Printf("Plan:       %s\n", plan)
	},
}

var plansCmd = &cobra.Command{
	Use:   "plans",
	Short: "Show available plans",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		result, err := client.Get("/api/plans")
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		plans, _ := result["plans"].([]interface{})
		fmt.Println("PLANS:")
		fmt.Printf("  %-8s %-12s %-18s %s\n", "ID", "Name", "Storage", "Price")
		fmt.Println("  " + "─────────────────────────────────────────────────")
		for _, p := range plans {
			plan := p.(map[string]interface{})
			id, _ := plan["id"].(string)
			name, _ := plan["name"].(string)
			storage := int64(plan["storage_bytes"].(float64))
			price := int(plan["price_cents"].(float64))

			priceStr := "free"
			if price > 0 {
				priceStr = fmt.Sprintf("$%d/mo", price/100)
			}
			fmt.Printf("  %-8s %-12s %-18s %s\n", id, name, formatBytes(storage), priceStr)
		}
		fmt.Println()
		fmt.Println("Upgrade: pidrive upgrade --plan <id>")
	},
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade your plan",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		plan, _ := cmd.Flags().GetString("plan")
		if plan == "" {
			fmt.Println("Usage: pidrive upgrade --plan <id>")
			fmt.Println("Run 'pidrive plans' to see available plans.")
			os.Exit(1)
		}

		_, err = client.Post("/api/upgrade", map[string]string{
			"plan": plan,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Upgraded to %s\n", plan)
	},
}

func init() {
	upgradeCmd.Flags().String("plan", "", "Plan to upgrade to (free, pro, team)")
}
