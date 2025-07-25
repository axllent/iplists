// Package adb is a package for interacting with the AbuseIPDb API.
package adb

import (
	"bufio"
	"encoding/json"
	"fmt"
	"iplists/cmd/internal/lib"
	"net/http"
	"os"
	"path"
	"sort"
	"time"
)

// Entry represents an entry in the AbuseIPDb database.
type Entry struct {
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

	// load all
	db := LoadExistingADBs(database, -1)
	// build a map for quick lookup
	existingIPs := make(map[string]Entry)
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
			existingIPs[ip] = Entry{
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
	updatedEntries := make([]Entry, 0, len(existingIPs))
	for _, k := range keys {
		updatedEntries = append(updatedEntries, existingIPs[k])
	}

	// write the updated entries to the database file
	file, err := os.Create(path.Clean(database))
	if err != nil {
		return fmt.Errorf("failed to create database file: %w", err)
	}
	defer func() { _ = file.Close() }()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // for pretty printing
	if err := encoder.Encode(updatedEntries); err != nil {
		return fmt.Errorf("failed to write to database file: %w", err)
	}

	fmt.Printf("Updated database with %d new IPs, removed %d expired IPs, total %d IPs.\n", added, removed, len(existingIPs))

	return nil
}

// LoadExistingADBs reads the existing IPs from the database file
// returning entries newer than N days.
func LoadExistingADBs(database string, days int) []Entry {
	entries := []Entry{}
	b, err := os.ReadFile(path.Clean(database))
	if err != nil {
		return entries
	}

	if err := json.Unmarshal(b, &entries); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// return all
	if days <= 0 {
		return entries
	}

	t := time.Now().Add(-time.Hour * 24 * time.Duration(days))

	cutoff := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	returnEntries := []Entry{}

	for _, entry := range entries {
		t, err := time.Parse(`2006-01-02`, entry.LastSeen)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if !t.Before(cutoff) {
			returnEntries = append(returnEntries, entry)
		}
	}

	return returnEntries
}
