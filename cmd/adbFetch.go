package cmd

import (
	"fmt"
	"iplists/cmd/internal/adb"
	"math/rand"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var adpPruneDays int

// adbFetchCmd represents the fetch command
var adbFetchCmd = &cobra.Command{
	Use:   "fetch <db-file>",
	Args:  cobra.ExactArgs(1),
	Short: "Update the database with the latest IPs from AbuseIPDb",
	Long: `This will update the local database with the latest IPs from AbuseIPDb.

By default IPs not seen in the last 100 days will be pruned from the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		keys, set := os.LookupEnv("ADB_KEY")
		if !set || keys == "" {
			fmt.Fprintln(os.Stderr, "ADB_KEY environment variable must be set to your AbuseIPDb API key")
			os.Exit(1)
		}

		parts := strings.Split(keys, ",")
		randomIndex := rand.Intn(len(parts)) // generate a random int
		key := parts[randomIndex]

		if err := adb.UpdateADBdb(key, args[0], adpPruneDays); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating AbuseIPDb database: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	adbCmd.AddCommand(adbFetchCmd)
	adbFetchCmd.Flags().IntVarP(&adpPruneDays, "days", "d", 100, "Prune stale IPs not seen in X days")
}
