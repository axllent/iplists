package cmd

import (
	"fmt"
	"iplists/cmd/internal/lib"
	"log"
	"net"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/projectdiscovery/mapcidr"
	"github.com/spf13/cobra"
)

var (
	aggregateOverwrite bool
	aggregateStatsOnly bool
)

// aggregateCmd represents the aggregate command
var aggregateCmd = &cobra.Command{
	Use:   "aggregate",
	Short: "Aggregate IPs/CIDRs into minimum IPs & subnets",
	Long:  `Aggregate IPs/CIDRs into minimum IPs & subnets.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		lines, err := lib.GetContents(path.Clean(args[0]))
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", args[0], err)
			return
		}

		var allCidrs []*net.IPNet

		// test if we have a cidr
		for _, cidr := range lines {
			if !strings.Contains(cidr, "/") {
				// if not a CIDR, try to parse as an IP
				ip := net.ParseIP(cidr)
				if ip == nil {
					log.Fatalf("Invalid IP or CIDR: %s\n", cidr)
				}

				// if it's a valid IP, convert it to a /32 CIDR
				if ip.To4() != nil {
					cidr = fmt.Sprintf("%s/32", ip.String())
				} else if ip.To16() != nil {
					cidr = fmt.Sprintf("%s/64", ip.String())
				} else {
					log.Fatalf("Invalid IP or CIDR: %s\n", cidr)
				}
			}

			_, pCidr, err := net.ParseCIDR(cidr)
			if err != nil {
				log.Fatalf("%s\n", err)
			}

			allCidrs = append(allCidrs, pCidr)
		}

		cCidrsIPV4, cCidrsIPV6 := mapcidr.CoalesceCIDRs(allCidrs)

		outputIPv4 := make([]string, 0, len(cCidrsIPV4))
		outputIPv6 := make([]string, 0, len(cCidrsIPV6))

		for _, cidr := range cCidrsIPV4 {
			if strings.HasSuffix(cidr.String(), "/32") {
				// if it's a /32 CIDR, print the IP only
				outputIPv4 = append(outputIPv4, strings.TrimSuffix(cidr.String(), "/32"))
			} else {
				outputIPv4 = append(outputIPv4, cidr.String())
			}
		}
		for _, cidr := range cCidrsIPV6 {
			outputIPv6 = append(outputIPv6, cidr.String())
		}

		sort.Strings(outputIPv4)
		sort.Strings(outputIPv6)

		if !aggregateOverwrite && !aggregateStatsOnly {
			for _, cidr := range outputIPv4 {
				fmt.Println(cidr)
			}
			for _, cidr := range outputIPv6 {
				fmt.Println(cidr)
			}
			return
		}

		if !aggregateStatsOnly {
			// write the updated entries to the cache file
			f, err := os.Create(path.Clean(args[0]))
			if err != nil {
				fmt.Printf("Failed to create cache file: %s\n", err.Error())
				os.Exit(1)
			}
			defer func() { _ = f.Close() }()

			for _, entry := range outputIPv4 {
				if _, err := fmt.Fprintln(f, entry); err != nil {
					fmt.Fprintf(os.Stderr, "Error writing to file %s: %v\n", args[1], err)
					os.Exit(1)
				}
			}

			for _, entry := range outputIPv6 {
				if _, err := fmt.Fprintln(f, entry); err != nil {
					fmt.Fprintf(os.Stderr, "Error writing to file %s: %v\n", args[1], err)
					os.Exit(1)
				}
			}
		}

		if len(lines) == 0 {
			fmt.Fprintf(os.Stderr, "No valid IPs found in file %s\n", args[1])
			os.Exit(1)
		}

		if len(lines) == len(outputIPv4)+len(outputIPv6) {
			fmt.Println("No aggregation needed, input and output are the same.")
			return
		}

		fmt.Printf(
			"Aggregated %s from %s IPs & CIDRs in %s\n",
			lib.NumberFormat(len(outputIPv4)+len(outputIPv6)),
			lib.NumberFormat(len(lines)),
			args[0],
		)
	},
}

func init() {
	rootCmd.AddCommand(aggregateCmd)

	aggregateCmd.Flags().BoolVarP(&aggregateOverwrite, "write", "w", false, "Overwrite file (default stdout)")
	aggregateCmd.Flags().BoolVarP(&aggregateStatsOnly, "stats", "s", false, "Show stats only, do not write to file")
}
