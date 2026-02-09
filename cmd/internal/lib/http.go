package lib

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Fetch will do a HTTP get request to the given URL and return the response body.
func Fetch(url string) ([]byte, error) {
	// Placeholder for HTTP GET implementation
	client := http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return []byte{}, err
	}

	response := resp.StatusCode
	if response != http.StatusOK {
		return []byte{}, fmt.Errorf("failed to fetch URL: %s, status code: %d", url, response)
	}

	defer resp.Body.Close()

	// Read the response body
	return io.ReadAll(resp.Body)
}
