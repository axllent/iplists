package cmd

import (
	"fmt"
	"iplists/cmd/internal/lib"
	"os"

	"github.com/spf13/cobra"
)

var (
	adbDays = 100
)

// adbCmd represents the adb command
var adbCmd = &cobra.Command{
	Use:   "adb <local_db> <output_file>",
	Short: "Update the AbuseIPDb database",
	Args:  cobra.ExactArgs(2),
	Long: `This maintains a list of IPs listed on the free AbuseIPDb database.

The free AbuseIPDb database only contains 10,000 entries, so this command will track IPs in this list
and retain IPs listed within the last 100 days.

ADB_KEY environment variable must be set with your AbuseIPDb API key.`,
	Run: func(cmd *cobra.Command, args []string) {
		key, set := os.LookupEnv("ADB_KEY")
		if !set || key == "" {
			fmt.Fprintln(os.Stderr, "ADB_KEY environment variable must be set to your AbuseIPDb API key")
			os.Exit(1)
		}

		if err := lib.UpdateADBdb(key, args[0], adbDays); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating AbuseIPDb database: %v\n", err)
			os.Exit(1)
		}

		entries := lib.LoadExistingADBs(args[0])
		if len(entries) == 0 {
			fmt.Println("No valid entries found in the local database.")
			return
		}

		f, err := os.OpenFile(args[1], os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening output file %s: %v\n", args[1], err)
			os.Exit(1)
		}
		defer func() { _ = f.Close() }()

		for _, entry := range entries {
			if _, err := fmt.Fprintln(f, entry.IP); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing to output file %s: %v\n", args[1], err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(adbCmd)

	adbCmd.Flags().IntVarP(&adbDays, "days", "d", 30, "Active in the last N days")
}
