package cmd

import (
	"fmt"
	"iplists/cmd/internal/o365"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// msIPsCmd represents the msIPs command
var o365IPsCmd = &cobra.Command{
	Use:   "o365-ips <output-file>",
	Short: "Fetch Microsoft IP ranges",
	Long: `Fetch and process Office365 IP ranges.
	
https://learn.microsoft.com/en-us/microsoft-365/enterprise/urls-and-ip-address-ranges?view=o365-worldwide

Currently this just writes the Microsoft Teams IPs to a file`,
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, _ []string) {
		data, err := o365.FetchO365IPs()
		if err != nil {
			fmt.Println("Error fetching Office365 IPs:", err)
			os.Exit(1)
		}

		ips, found := data["Microsoft Teams"]
		list := make(map[string]bool)
		uniqueIPs := []string{}

		if found {
			for _, ip := range ips {
				if _, found := list[ip]; !found {
					list[ip] = true
					uniqueIPs = append(uniqueIPs, ip)
				}
			}
		}

		if len(uniqueIPs) > 0 {
			// write to file
			outputFile := os.Args[2]
			file, err := os.Create(filepath.Clean(outputFile))
			if err != nil {
				fmt.Println("Error creating output file:", err)
				os.Exit(1)
			}
			defer file.Close()

			for _, ip := range uniqueIPs {
				_, err := file.WriteString(ip + "\n")
				if err != nil {
					fmt.Println("Error writing to output file:", err)
					os.Exit(1)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(o365IPsCmd)
}
