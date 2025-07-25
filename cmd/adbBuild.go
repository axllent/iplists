package cmd

import (
	"fmt"
	"iplists/cmd/internal/adb"
	"os"

	"github.com/spf13/cobra"
)

var adbDays = 30

// adbBuildCmd represents the build command
var adbBuildCmd = &cobra.Command{
	Use:   "build <db-file> <output-file>",
	Args:  cobra.ExactArgs(2),
	Short: "Build a list of IPs from AbuseIPDb database",
	Long: `This command builds a list of IPs from the AbuseIPDb database.
	
It will read the local database and output a list of IPs that are currently listed,
active in the last N days.`,
	Run: func(cmd *cobra.Command, args []string) {
		entries := adb.LoadExistingADBs(args[0], adbDays)
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

		ips := 0
		for _, entry := range entries {
			if _, err := fmt.Fprintln(f, entry.IP); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing to output file %s: %v\n", args[1], err)
				os.Exit(1)
			}
			ips++
		}

		if adbDays <= 0 {
			fmt.Printf("Wrote %d entries to %s\n", ips, args[1])
			return
		}

		fmt.Printf("Wrote %d ips active in the last %d days to %s\n", ips, adbDays, args[1])
	},
}

func init() {
	adbCmd.AddCommand(adbBuildCmd)
	adbBuildCmd.Flags().IntVarP(&adbDays, "days", "d", 30, "Active in the last N days")
}
