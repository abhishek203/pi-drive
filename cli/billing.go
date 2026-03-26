package cli

import (
	"fmt"
	"os"
)

func runUsage(args []string) {
	parsed, err := parseCommandArgs(args, nil)
	if err != nil {
		fatalf("%v", err)
	}
	if len(parsed.args) != 0 {
		fmt.Println("Usage: pidrive usage")
		os.Exit(1)
	}

	client, err := NewClient()
	if err != nil {
		fatalf("%v", err)
	}

	result, err := client.Get("/api/usage")
	if err != nil {
		fatalf("%v", err)
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
}

func runPlans(args []string) {
	parsed, err := parseCommandArgs(args, nil)
	if err != nil {
		fatalf("%v", err)
	}
	if len(parsed.args) != 0 {
		fmt.Println("Usage: pidrive plans")
		os.Exit(1)
	}

	client, err := NewClient()
	if err != nil {
		fatalf("%v", err)
	}

	result, err := client.Get("/api/plans")
	if err != nil {
		fatalf("%v", err)
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
}

func runUpgrade(args []string) {
	parsed, err := parseCommandArgs(args, map[string]flagType{"plan": stringFlag})
	if err != nil {
		fatalf("%v", err)
	}

	plan := parsed.String("plan", "")
	if plan == "" || len(parsed.args) != 0 {
		fmt.Println("Usage: pidrive upgrade --plan <id>")
		fmt.Println("Run 'pidrive plans' to see available plans.")
		os.Exit(1)
	}

	client, err := NewClient()
	if err != nil {
		fatalf("%v", err)
	}
	if _, err := client.Post("/api/upgrade", map[string]string{"plan": plan}); err != nil {
		fatalf("%v", err)
	}

	fmt.Printf("✓ Upgraded to %s\n", plan)
}
