package o365

import (
	"encoding/json"
	"iplists/cmd/internal/lib"
)

var (
	// https://learn.microsoft.com/en-us/microsoft-365/enterprise/urls-and-ip-address-ranges?view=o365-worldwide
	o365URL = "https://endpoints.office.com/endpoints/worldwide?clientrequestid=b10c5ed1-bad1-445f-b386-b919946339a7"
)

// FetchO365IPs fetches and processes Office365 IP ranges.
func FetchO365IPs() (map[string][]string, error) {
	output := make(map[string][]string)

	data, err := lib.Fetch(o365URL)
	if err != nil {
		return output, err
	}

	var r = jsonData{}

	// parse data
	if err := json.Unmarshal(data, &r); err != nil {
		return output, err
	}

	for _, d := range r {
		if len(d.Ips) == 0 {
			continue
		}

		data := []string{}
		if _, ok := output[d.ServiceAreaDisplayName]; ok {
			data = output[d.ServiceAreaDisplayName]
		}

		for _, ip := range d.Ips {
			if lib.ValidAddress(ip) {
				data = append(data, ip)
			}
		}

		output[d.ServiceAreaDisplayName] = data
	}

	// Placeholder for fetching and processing Office365 IP ranges
	return output, nil
}
