package cmd

import (
	"fmt"
	"iplists/cmd/internal/lib"
	"net"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// pruneCmd represents the prune command
var pruneCmd = &cobra.Command{
	Use:   "prune <this_list> <with_this_list>",
	Args:  cobra.ExactArgs(2),
	Short: "Prune a list of IPs or CIDRs from another list",
	Long: `Compares two lists of IPs or CIDRs and removes entries from "this_list"
which are present in "with_list_list".`,
	Run: func(cmd *cobra.Command, args []string) {

		dst, err := lib.GetContents(args[0])
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error reading source file %s: %v\n", args[0], err)
			os.Exit(1)
			return
		}

		fromList, err := lib.GetContents(args[1])
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error reading destination file %s: %v\n", args[1], err)
			os.Exit(1)
			return
		}

		// exact matches
		fromListExact := make(map[string]bool)
		// map prefix to a slice of *net.IPNet for fast lookup
		fromListCIDR := make(map[string][]*net.IPNet)

		for _, entry := range fromList {
			if entry == "" {
				continue
			}
			fromListExact[entry] = true

			if strings.Contains(entry, "/") {
				// if the entry contains a '/', treat it as a CIDR
				if _, cidr, err := net.ParseCIDR(entry); err == nil {
					prefix := cidrPrefix(entry)
					arr, ok := fromListCIDR[prefix]
					if !ok {
						arr = []*net.IPNet{}
					}
					arr = append(arr, cidr)
					fromListCIDR[prefix] = arr
				} else {
					fmt.Fprintf(cmd.ErrOrStderr(), "Invalid CIDR in this_list: %s\n", entry)
				}
				continue
			}
		}

		newList := []string{}
		removed := 0

		for _, entry := range dst {
			if entry == "" {
				continue
			}

			toScan := entry

			if strings.Contains(toScan, "/") {
				_, ok := fromListExact[toScan]
				if ok {
					removed++
					continue
				}

				toScan = strings.Split(entry, "/")[0]
			}

			prefix := cidrPrefix(toScan)
			ip := net.ParseIP(toScan)

			found := false

			arr, ok := fromListCIDR[prefix]
			if ok {
				for _, cidr := range arr {
					if cidr.Contains(ip) {
						removed++
						found = true
						continue
					}
				}
			}

			if !found {
				newList = append(newList, entry)
			}
		}

		f, err := os.OpenFile(args[0], os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file %s: %v\n", args[0], err)
			os.Exit(1)
		}
		defer func() { _ = f.Close() }()

		for _, entry := range newList {
			if _, err := fmt.Fprintln(f, entry); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing to output file %s: %v\n", args[1], err)
				os.Exit(1)
			}
		}

		fmt.Printf("Pruned %d entries from %s\n", removed, args[0])

	},
}

func init() {
	rootCmd.AddCommand(pruneCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pruneCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pruneCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// CIDRPrefix will return the first two parts of an IP, used for
func cidrPrefix(ip string) string {
	if strings.Contains(ip, ":") {
		//ipv6
		parts := strings.Split(ip, ":")
		return fmt.Sprintf("%s:%s", parts[0], parts[1])
	}

	// ipv4
	parts := strings.Split(ip, ".")
	return fmt.Sprintf("%s.%s", parts[0], parts[1])
}
