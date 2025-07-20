package lib

import (
	"bufio"
	"os"
	"strings"
)

// GetContents reads a file and returns its contents as a slice of strings,
func GetContents(file string) ([]string, error) {
	// Open the file
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	// Read the file line by line
	var contents []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			contents = append(contents, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return contents, nil
}
