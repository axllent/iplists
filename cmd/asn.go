package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/oschwald/maxminddb-golang"
	"github.com/spf13/cobra"
)

var asnDBPath string
var asnIDsOnly bool

// asnCmd represents the asn command
var asnCmd = &cobra.Command{
	Use:   "asn <search>",
	Short: "Search for ASN IDs by name in the GeoLite2 ASN database",
	Long: `The asn command searches the GeoLite2 ASN MaxMind database for entries
matching the given search string (case-insensitive partial match).

Results are printed as CSV: ASN,Name`,
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		search := strings.ToLower(args[0])

		db, err := maxminddb.Open(asnDBPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		type asnRecord struct {
			AutonomousSystemNumber       uint   `maxminddb:"autonomous_system_number"`
			AutonomousSystemOrganization string `maxminddb:"autonomous_system_organization"`
		}

		seen := make(map[uint]bool)

		networks := db.Networks(maxminddb.SkipAliasedNetworks)
		for networks.Next() {
			var record asnRecord
			_, err := networks.Network(&record)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading record: %v\n", err)
				os.Exit(1)
			}

			if record.AutonomousSystemNumber == 0 || seen[record.AutonomousSystemNumber] {
				continue
			}

			if strings.Contains(strings.ToLower(record.AutonomousSystemOrganization), search) {
				seen[record.AutonomousSystemNumber] = true
				if !asnIDsOnly {
					fmt.Printf("AS%d,%s\n", record.AutonomousSystemNumber, record.AutonomousSystemOrganization)
				}
			}
		}

		if err := networks.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error iterating networks: %v\n", err)
			os.Exit(1)
		}

		if asnIDsOnly {
			ids := make([]string, 0, len(seen))
			for id := range seen {
				ids = append(ids, fmt.Sprintf("AS%d", id))
			}
			fmt.Println(strings.Join(ids, " "))
		}
	},
}

func init() {
	rootCmd.AddCommand(asnCmd)
	asnCmd.Flags().StringVarP(&asnDBPath, "db", "d", "/usr/share/GeoIP/GeoLite2-ASN.mmdb", "Path to GeoLite2 ASN database")
	asnCmd.Flags().BoolVarP(&asnIDsOnly, "ids", "i", false, "Print only ASN IDs space-separated")
}
