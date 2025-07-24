// Package lib is a general library
package lib

import (
	"net"
	"strings"
)

// ValidAddress checks if the given IP or CIDR is valid and not a private address.
func ValidAddress(ip string) bool {
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
