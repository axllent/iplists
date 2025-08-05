// Package lib is a general library
package lib

import (
	"net"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// ValidAddress checks if the given IP or CIDR is valid and not a private address.
// @see https://en.wikipedia.org/wiki/Reserved_IP_addresses
func ValidAddress(ip string) bool {
	if strings.Contains(ip, "/") {
		// parse as a CIDR notation
		parsedIP, _, err := net.ParseCIDR(ip)
		if err != nil {
			return false
		}
		// Go seems to miss a few important entries, namely 127.* and 0.*
		if parsedIP == nil || parsedIP.IsPrivate() || strings.HasPrefix(ip, "127.") || strings.HasPrefix(ip, "0.") {
			return false
		}
	} else {
		// Check if it's a CIDR notation
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil || parsedIP.IsPrivate() || strings.HasPrefix(ip, "127.") || strings.HasPrefix(ip, "0.") {
			return false
		}
	}

	return true
}

// NumberFormat formats a number using the English locale.
func NumberFormat(d int) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%d", d)
}
