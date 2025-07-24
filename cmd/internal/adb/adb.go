// Package adb is a package for interacting with the AbuseIPDb API.
package adb

import (
	"bufio"
	"encoding/json"
	"fmt"
	"iplists/cmd/internal/lib"
	"net/http"
	"os"
	"sort"
	"time"
)

// ADBEntry represents an entry in the AbuseIPDb database.
type ADBEntry struct {
	IP       string `json:"ip"`
	LastSeen string `json:"last_seen"`
}

// UpdateADBdb fetches the AbuseIPDb blacklist and updates the local database.
func UpdateADBdb(key, database string, days int) error {
	req, _ := http.NewRequest("GET", "https://api.abuseipdb.com/api/v2/blacklist", nil)
	req.Header.Set("Key", key)
	req.Header.Set("Accept", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	scanner := bufio.NewScanner(resp.Body)

	listIPs := []string{}

	for scanner.Scan() {
		i := scanner.Text()
		if lib.ValidAddress(i) {
			listIPs = append(listIPs, i)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if len(listIPs) == 0 {
		return fmt.Errorf("no valid IPs found in the response")
	}

	db := LoadExistingADBs(database)
	// build a map for quick lookup
	existingIPs := make(map[string]ADBEntry)
	for _, entry := range db {
		existingIPs[entry.IP] = entry
	}

	added := 0

	for _, ip := range listIPs {
		now := time.Now().UTC().Format(`2006-01-02`)
		if entry, exists := existingIPs[ip]; exists {
			// If the IP already exists, update the last seen time
			entry.LastSeen = now
			existingIPs[ip] = entry
		} else {
			// If the IP does not exist, add it with the current time
			existingIPs[ip] = ADBEntry{
				IP:       ip,
				LastSeen: now,
			}
			added++
		}
	}

	removed := 0

	// remove expired entries
	for ip, entry := range existingIPs {
		t, err := time.Parse(`2006-01-02`, entry.LastSeen)
		if err != nil {
			fmt.Println(err)
			fmt.Println("")
			continue
		}

		if time.Since(t).Hours() > float64(days*24) {
			delete(existingIPs, ip)
			removed++
		}
	}

	keys := make([]string, 0, len(existingIPs))

	for k := range existingIPs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// build the updated list in order of IP addresses
	updatedEntries := make([]ADBEntry, 0, len(existingIPs))
	for _, k := range keys {
		updatedEntries = append(updatedEntries, existingIPs[k])
	}

	// write the updated entries to the database file
	file, err := os.Create(database)
	if err != nil {
		return fmt.Errorf("failed to create database file: %w", err)
	}
	defer func() { _ = file.Close() }()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // for pretty printing
	if err := encoder.Encode(updatedEntries); err != nil {
		return fmt.Errorf("failed to write to database file: %w", err)
	}

	fmt.Printf("[AbuseIPDB] Updated database with %d new entries, removed %d expired entries, total (%d entries).\n", added, removed, len(existingIPs))

	return nil
}

// LoadExistingADBs reads the existing IPs from the database file.
func LoadExistingADBs(database string) []ADBEntry {
	entries := []ADBEntry{}
	b, err := os.ReadFile(database)
	if err != nil {
		return entries
	}

	if err := json.Unmarshal(b, &entries); err != nil {
		fmt.Printf("[AbuseIPDB] %s\n", err.Error())
	}

	return entries
}
