package cmd

import (
	"bufio"
	"fmt"
	"iplists/cmd/internal/lib"
	"net"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/projectdiscovery/mapcidr"
	"github.com/spf13/cobra"
)

// asnIPsCmd represents the asn-ips command
var asnIPsCmd = &cobra.Command{
	Use:   "asn-ips <asn> [asn...]",
	Short: "Fetch and aggregate IP ranges for one or more ASNs via RADB whois",
	Long: `The asn-ips command queries whois.radb.net for route and route6 entries
for the given ASN numbers, validates the results, aggregates them into the
minimum set of IPs and CIDRs, and prints them to stdout.

ASN arguments can be specified with or without the "AS" prefix (e.g. AS13414 or 13414).`,
	Args: cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		routeMatch := regexp.MustCompile(`^route6?:\s+(\S+)`)

		var allCIDRs []*net.IPNet

		for _, asn := range args {
			asn = strings.TrimSpace(asn)
			if !strings.HasPrefix(strings.ToUpper(asn), "AS") {
				asn = "AS" + asn
			} else {
				asn = "AS" + strings.TrimPrefix(strings.ToUpper(asn), "AS")
			}

			out, err := exec.Command("whois", "-h", "whois.radb.net", "--", "-i", "origin", asn).Output()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error querying RADB for %s: %v\n", asn, err)
				continue
			}

			scanner := bufio.NewScanner(strings.NewReader(string(out)))
			for scanner.Scan() {
				line := scanner.Text()
				m := routeMatch.FindStringSubmatch(line)
				if m == nil {
					continue
				}

				entry := strings.TrimSpace(m[1])
				if !lib.ValidAddress(entry) {
					continue
				}

				if !strings.Contains(entry, "/") {
					ip := net.ParseIP(entry)
					if ip == nil {
						continue
					}
					if ip.To4() != nil {
						entry = ip.String() + "/32"
					} else {
						entry = ip.String() + "/128"
					}
				}

				_, cidr, err := net.ParseCIDR(entry)
				if err != nil {
					continue
				}
				allCIDRs = append(allCIDRs, cidr)
			}
		}

		if len(allCIDRs) == 0 {
			fmt.Fprintln(os.Stderr, "No valid routes found")
			os.Exit(1)
		}

		ipv4, ipv6 := mapcidr.CoalesceCIDRs(allCIDRs)

		output := make([]string, 0, len(ipv4)+len(ipv6))
		for _, cidr := range ipv4 {
			s := cidr.String()
			if before, ok := strings.CutSuffix(s, "/32"); ok {
				s = before
			}
			output = append(output, s)
		}
		for _, cidr := range ipv6 {
			output = append(output, cidr.String())
		}

		sort.Strings(output)

		for _, entry := range output {
			fmt.Println(entry)
		}
	},
}

func init() {
	rootCmd.AddCommand(asnIPsCmd)
}
