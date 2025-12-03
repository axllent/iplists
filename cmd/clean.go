package cmd

import (
	"bufio"
	"fmt"
	"iplists/cmd/internal/lib"
	"os"
	"regexp"

	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean < <list> || cat <file> | iplist clean",
	Short: "Clean command to filter and validate IPs and CIDRs",
	Long: `The clean command reads lines of text from standard input and outputs valid IPs and CIDRs.

IPs should be piped to this command, and it will filter out invalid entries, including private addresses.`,
	Run: func(_ *cobra.Command, _ []string) {
		scanner := bufio.NewScanner(os.Stdin)
		ipMatch := regexp.MustCompile(`^(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}|[0-9a-fA-F:]{6,})(\/\d{1,2})?`)

		for scanner.Scan() {
			line := scanner.Text()
			if ipMatch.MatchString(line) && lib.ValidAddress(line) {
				fmt.Println(line)
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
