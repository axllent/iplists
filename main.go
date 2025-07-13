// Package main provides a simple command-line application that reads lines of text from
// standard input and output valid IPs and CIDRs.
package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	ipMatch := regexp.MustCompile(`^(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}|[0-9a-fA-F:]{6,})(\/\d{1,2})?`)

	for scanner.Scan() {
		line := scanner.Text()
		if ipMatch.MatchString(line) && valid(line) {
			fmt.Println(line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
		os.Exit(1)
	}
}

// valid checks if the given IP or CIDR is valid and not a private address.
func valid(ip string) bool {
	if strings.Contains(ip, "/") {
		// parse as a CIDR notation
		parsedIP, _, err := net.ParseCIDR(ip)
		if err != nil {
			return false
		}

		if parsedIP == nil || parsedIP.IsPrivate() {
			return false
		}
	} else {
		// Check if it's a CIDR notation
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil || parsedIP.IsPrivate() {
			return false
		}
	}

	return true
}
