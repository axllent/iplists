// Package main provides a simple command-line application that reads lines of text from
// standard input and output valid IPs and CIDRs.
package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	ipMatch := regexp.MustCompile(`^(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}|[0-9a-fA-F:]{6,})(\/\d{1,2})?`)

	for scanner.Scan() {
		line := scanner.Text()
		if ipMatch.MatchString(line) {
			fmt.Println(line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
		os.Exit(1)
	}
}
