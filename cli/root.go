package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pidrive",
	Short: "Google Drive for AI Agents",
	Long:  `pidrive — Private file storage for AI agents. Files on S3, agents use ls, grep, cat. Share via URLs.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(registerCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(whoamiCmd)
	rootCmd.AddCommand(mountCmd)
	rootCmd.AddCommand(unmountCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(shareCmd)
	rootCmd.AddCommand(sharedCmd)
	rootCmd.AddCommand(revokeCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(activityCmd)
	rootCmd.AddCommand(trashCmd)
	rootCmd.AddCommand(restoreCmd)
	rootCmd.AddCommand(usageCmd)
	rootCmd.AddCommand(plansCmd)
	rootCmd.AddCommand(upgradeCmd)
}
