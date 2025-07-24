package cmd

import (
	"github.com/spf13/cobra"
)

// adbCmd represents the adb command
var adbCmd = &cobra.Command{
	Use:   "adb",
	Short: "Update the AbuseIPDb database",
	// Args: cobra.NoArgs,
	Long: `This maintains a list of IPs listed on the free AbuseIPDb database.

The free AbuseIPDb database only contains 10,000 entries, so this command will track IPs in this list
and retain IPs listed within the last 100 days.

ADB_KEY environment variable must be set with your AbuseIPDb API key.`,
	// Run: func(cmd *cobra.Command, args []string) {
	// },
}

func init() {
	rootCmd.AddCommand(adbCmd)
}
